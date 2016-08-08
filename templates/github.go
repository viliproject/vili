package templates

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/airware/vili/config"
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

func (s *githubService) getContents(env, branch, subPath string) (*github.RepositoryContent, []*github.RepositoryContent, *github.Response, error) {
	envContentsPath, ok := s.config.EnvContentsPaths[env]
	if !ok {
		envContentsPath = s.config.EnvContentsPaths[config.GetString(config.DefaultEnv)]
	}
	var opts *github.RepositoryContentGetOptions
	if branch != "" {
		opts = &github.RepositoryContentGetOptions{Ref: branch}
	}
	fileContent, directoryContent, response, err := s.client.Repositories.GetContents(s.config.Owner, s.config.Repo, fmt.Sprintf(envContentsPath, subPath), opts)
	if _, ok := err.(*github.ErrorResponse); ok && branch != "" {
		// Fall back to the default branch
		return s.getContents(env, "", subPath)
	}
	return fileContent, directoryContent, response, err
}

// Deployments returns a list of deployments for the given environment
func (s *githubService) Deployments(env, branch string) ([]string, error) {
	_, directoryContent, _, err := s.getContents(env, branch, "deployments")
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

// Deployment returns a deployment for the given environment
func (s *githubService) Deployment(env, branch, name string) (Template, error) {
	fileContent, _, _, err := s.getContents(env, branch, "deployments/"+name+".yaml")
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
func (s *githubService) Pods(env, branch string) ([]string, error) {
	_, directoryContent, _, err := s.getContents(env, branch, "pods")
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
func (s *githubService) Pod(env, branch, name string) (Template, error) {
	fileContent, _, _, err := s.getContents(env, branch, "pods/"+name+".yaml")
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

// Environment returns an environment template for the given branch
func (s *githubService) Environment(branch string) (Template, error) {
	fileContent, _, _, err := s.getContents("", branch, "environment.yaml")
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
