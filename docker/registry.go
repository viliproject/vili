package docker

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/docker/distribution"
	"github.com/docker/distribution/context"
	"github.com/docker/distribution/reference"
	"github.com/docker/distribution/registry/client"
	"github.com/docker/distribution/registry/client/auth"
	"github.com/docker/distribution/registry/client/transport"
)

// RegistryConfig is the registry service configuration
type RegistryConfig struct {
	BaseURL         string
	Username        string
	Password        string
	Namespace       string
	BranchDelimiter string
}

// RegistryService is an implementation of the docker Service interface
// It fetches docker images
type RegistryService struct {
	config *RegistryConfig
}

// InitRegistry initializes the docker registry service
func InitRegistry(c *RegistryConfig) error {
	service = &RegistryService{
		config: c,
	}
	return nil
}

// GetRepository implements the Service interface
func (s *RegistryService) GetRepository(repo string, withBranches bool) ([]*Image, error) {
	var waitGroup sync.WaitGroup
	imagesChan := make(chan getImagesResult, len(branches))

	if withBranches {
		for _, branch := range branches {
			if branch != "master" {
				waitGroup.Add(1)
				go func(branch string) {
					defer waitGroup.Done()
					images, err := s.getImagesForBranch(repo, branch)
					imagesChan <- getImagesResult{images: images, err: err}
				}(branch)
			}
		}
	}

	images, err := s.getImagesForBranch(repo, "master")
	if err != nil {
		return nil, err
	}

	waitGroup.Wait()
	close(imagesChan)
	for result := range imagesChan {
		if result.err != nil {
			return nil, result.err
		}
		images = append(images, result.images...)
	}

	sortByLastModified(images)
	return images, nil
}

// GetTag implements the Service interface
func (s *RegistryService) GetTag(repo, branch, tag string) (string, error) {
	repository, err := s.getRepositoryForBranch(repo, branch)
	if err != nil {
		return "", err
	}

	desc, err := repository.Tags(context.Background()).Get(context.Background(), tag)
	if err != nil {
		return "", err
	}

	return desc.Digest.String(), nil
}

// FullName implements the Service interface
func (s *RegistryService) FullName(repo, branch, tag string) (string, error) {
	if s.config.Namespace != "" {
		repo = s.config.Namespace + "/" + repo
	}
	if branch != "master" {
		repo += s.config.BranchDelimiter + strings.ToLower(branch)
	}
	return s.config.BaseURL + "/" + repo + ":" + tag, nil
}

func (s *RegistryService) getImagesForBranch(repoName, branchName string) ([]*Image, error) {
	repo, err := s.getRepositoryForBranch(repoName, branchName)
	if err != nil {
		return nil, err
	}

	tags, err := repo.Tags(context.Background()).All(context.Background())
	if err != nil {
		return nil, err
	}

	var images []*Image
	for _, tag := range tags {
		image := &Image{
			Tag:    tag,
			Branch: branchName,
		}
		sepIndex := strings.LastIndex(tag, "-")
		if sepIndex != -1 {
			dateComponent, shaComponent := tag[:sepIndex], tag[sepIndex+1:]
			unixSecs, err := strconv.ParseInt(dateComponent, 10, 0)
			if err != nil {
				continue
			}
			image.Revision = shaComponent
			image.LastModified = time.Unix(unixSecs, 0)
		}
		images = append(images, image)
	}
	return images, nil
}

func (s *RegistryService) getRepositoryForBranch(repoName, branchName string) (distribution.Repository, error) {
	if s.config.Namespace != "" {
		repoName = s.config.Namespace + "/" + repoName
	}
	if branchName != "master" {
		repoName += s.config.BranchDelimiter + strings.ToLower(branchName)
	}
	repoNameRef, err := reference.ParseNamed(repoName)
	if err != nil {
		return nil, err
	}

	credentialStore := &basicCredentialStore{
		Username: s.config.Username,
		Password: s.config.Password,
	}

	challengeManager := auth.NewSimpleChallengeManager()
	resp, err := http.Get(s.config.BaseURL + "/v2/")
	if err != nil {
		return nil, err
	}
	if err := challengeManager.AddResponse(resp); err != nil {
		return nil, err
	}

	transport := transport.NewTransport(http.DefaultTransport, auth.NewAuthorizer(
		challengeManager,
		auth.NewTokenHandler(http.DefaultTransport, credentialStore, repoName, "pull"),
		auth.NewBasicHandler(credentialStore),
	))

	repo, err := client.NewRepository(context.Background(), repoNameRef, s.config.BaseURL, transport)
	if err != nil {
		return nil, err
	}

	return repo, nil
}

// basicCredentialStore implements the distribution auth.CredentialStore interface
// for use with a single registry.
type basicCredentialStore struct {
	Username string
	Password string
}

func (cs *basicCredentialStore) Basic(u *url.URL) (string, string) {
	return cs.Username, cs.Password
}

func (cs *basicCredentialStore) RefreshToken(u *url.URL, service string) string {
	return ""
}

func (cs *basicCredentialStore) SetRefreshToken(realm *url.URL, service, token string) {
}
