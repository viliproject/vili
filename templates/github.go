package templates

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// GithubConfig is the configuration for the github client
type GithubConfig struct {
	Token            string
	Owner            string
	Repo             string
	EnvContentsPaths map[string]string
}

type githubService struct {
	config *GithubConfig
	client *github.Client
}

// InitGithub initializes the github templates service with the given config
func InitGithub(config *GithubConfig) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.Token},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	service = &githubService{
		config: config,
		client: github.NewClient(tc),
	}
}

// Controllers returns a list of controllers for the given environment
func (s *githubService) Controllers(env string) ([]string, error) {
	path := fmt.Sprintf(s.config.EnvContentsPaths[env], "controllers")
	_, directoryContent, _, err := s.client.Repositories.GetContents(s.config.Owner, s.config.Repo, path, nil)
	if err != nil {
		return nil, err
	}
	var controllers []string
	for _, content := range directoryContent {
		parts := strings.Split(*content.Name, ".")
		if len(parts) != 2 || parts[1] != "yaml" {
			continue
		}
		controllers = append(controllers, parts[0])
	}
	return controllers, nil
}

// Controller returns a list of controllers for the given environment
func (s *githubService) Controller(env, name string) (Template, error) {
	path := fmt.Sprintf(s.config.EnvContentsPaths[env], "controllers/"+name+".yaml")
	fileContent, _, _, err := s.client.Repositories.GetContents(s.config.Owner, s.config.Repo, path, nil)
	if err != nil {
		return "", err
	}
	if fileContent.DownloadURL == nil {
		return "", fmt.Errorf("no download url in github file response")
	}

	resp, err := http.Get(*fileContent.DownloadURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return Template(body), nil
}

// Pods returns a list of pods for the given environment
func (s *githubService) Pods(env string) ([]string, error) {
	path := fmt.Sprintf(s.config.EnvContentsPaths[env], "pods")
	_, directoryContent, _, err := s.client.Repositories.GetContents(s.config.Owner, s.config.Repo, path, nil)
	if err != nil {
		if errResp, ok := err.(*github.ErrorResponse); ok {
			if errResp.Response.StatusCode == 404 {
				return nil, nil
			}
		}
		return nil, err
	}
	var pods []string
	for _, content := range directoryContent {
		parts := strings.Split(*content.Name, ".")
		if len(parts) != 2 || parts[1] != "yaml" {
			continue
		}
		pods = append(pods, parts[0])
	}
	return pods, nil
}

// Pod returns a list of pods for the given environment
func (s *githubService) Pod(env, name string) (Template, error) {
	path := fmt.Sprintf(s.config.EnvContentsPaths[env], "pods/"+name+".yaml")
	fileContent, _, _, err := s.client.Repositories.GetContents(s.config.Owner, s.config.Repo, path, nil)
	if err != nil {
		return "", err
	}
	if fileContent.DownloadURL == nil {
		return "", fmt.Errorf("no download url in github file response")
	}

	resp, err := http.Get(*fileContent.DownloadURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return Template(body), nil
}

// Variables returns a list of variabless for the given environment
func (s *githubService) Variables(env string) (map[string]string, error) {
	path := fmt.Sprintf(s.config.EnvContentsPaths[env], "variables/"+env+".json")
	fileContent, _, _, err := s.client.Repositories.GetContents(s.config.Owner, s.config.Repo, path, nil)
	if err != nil {
		return nil, err
	}
	if fileContent.DownloadURL == nil {
		return nil, fmt.Errorf("no download url in github file response")
	}

	resp, err := http.Get(*fileContent.DownloadURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	variables := map[string]string{
		"ENV": env,
	}
	err = json.Unmarshal(body, &variables)
	if err != nil {
		return nil, err
	}
	return variables, nil
}
