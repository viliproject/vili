package docker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/airware/vili/log"
)

// QuayConfig is the quay service configuration
type QuayConfig struct {
	Token     string
	Namespace string
}

// QuayService is an implementation of the docker Service interface
// It fetches docker image
type QuayService struct {
	config *QuayConfig
}

// InitQuay initializes the quay docker service
func InitQuay(c *QuayConfig) error {
	service = &QuayService{
		config: c,
	}
	return nil
}

// GetRepository implements the Service interface
func (s *QuayService) GetRepository(repo string, withBranches bool) ([]*Image, error) {
	var waitGroup sync.WaitGroup
	tagImageIDsChan := make(chan getTagImageIDsResult, len(branches))
	if withBranches {
		for _, branch := range branches {
			waitGroup.Add(1)
			go func(branch string) {
				defer waitGroup.Done()
				s.getTagImageIDsIntoChan(repo, branch, tagImageIDsChan)
			}(branch)
		}
	}
	repoResp := &QuayRepositoryResponse{}
	err := s.makeGetRequest(
		fmt.Sprintf("repository/%s/%s", s.config.Namespace, repo),
		nil,
		repoResp,
	)
	if err != nil {
		return nil, err
	}
	waitGroup.Wait()
	close(tagImageIDsChan)
	imageBranches := make(map[string]string)
	for result := range tagImageIDsChan {
		if result.err != nil {
			return nil, result.err
		}
		for _, imageID := range result.imageIDs {
			imageBranches[imageID] = result.tag
		}
	}

	var images []*Image
	for tag, tagData := range repoResp.Tags {
		isBranch := false
		for _, branch := range branches {
			if tag == branch {
				isBranch = true
				break
			}
		}
		if isBranch {
			continue
		}
		lastModified, err := time.Parse(time.RFC1123Z, tagData["last_modified"].(string))
		if err != nil {
			log.Error(err)
			continue
		}
		imageID := tagData["image_id"].(string)
		size, sizeOK := tagData["size"].(float64)
		if !sizeOK {
			continue
		}
		images = append(images, &Image{
			ID:           imageID,
			Size:         int(size),
			Tag:          tag,
			Branch:       imageBranches[imageID],
			LastModified: lastModified,
		})
	}
	sortByLastModified(images)
	return images, nil
}

// GetTagImageIDs implements the Service interface
func (s *QuayService) GetTagImageIDs(repo, tag string) ([]string, error) {
	tagImagesResp := &QuayTagImagesResponse{}
	err := s.makeGetRequest(
		fmt.Sprintf("repository/%s/%s/tag/", s.config.Namespace, repo),
		&url.Values{
			"specificTag": []string{tag},
			"limit":       []string{"100"},
		},
		tagImagesResp,
	)
	if err != nil {
		return nil, err
	}
	var imageIDs []string
	for _, tagData := range tagImagesResp.Tags {
		imageIDs = append(imageIDs, tagData["docker_image_id"].(string))
	}
	return imageIDs, nil
}

type getTagImageIDsResult struct {
	tag      string
	imageIDs []string
	err      error
}

func (s *QuayService) getTagImageIDsIntoChan(repo, tag string, doneChan chan<- getTagImageIDsResult) {
	imageIDs, err := s.GetTagImageIDs(repo, tag)
	doneChan <- getTagImageIDsResult{tag: tag, imageIDs: imageIDs, err: err}
}

func (s *QuayService) makeGetRequest(path string, params *url.Values, dest interface{}) error {
	urlStr := fmt.Sprintf("https://quay.io/api/v1/%s", path)
	if params != nil {
		urlStr += "?" + params.Encode()
	}
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.config.Token))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusNotFound {
		return &NotFoundError{}
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Failed Quay request: %s", string(body))
	}
	json.Unmarshal(body, dest)
	return nil
}

// QuayRepositoryResponse is the json response for repository requests
type QuayRepositoryResponse struct {
	Tags map[string]map[string]interface{} `json:"tags"`
}

// QuayTagImagesResponse is the json response for tag images requests
type QuayTagImagesResponse struct {
	Tags []map[string]interface{} `json:"tags"`
}
