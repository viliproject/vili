package templates

import (
	"fmt"
	"strings"

	"github.com/airware/vili/config"
	"github.com/airware/vili/git"
)

// GitConfig is the configuration for the git tempaltes service
type GitConfig struct {
	EnvContentsPaths map[string]string
}

type gitService struct {
	config *GitConfig
}

// InitGit initializes the git templates service with the global git service
func InitGit(config *GitConfig) {
	service = &gitService{
		config: config,
	}
}

func (s *gitService) resolvePath(env, subPath string) string {
	envContentsPath, ok := s.config.EnvContentsPaths[env]
	if !ok {
		envContentsPath = s.config.EnvContentsPaths[config.GetString(config.DefaultEnv)]
	}
	return fmt.Sprintf(envContentsPath, subPath)
}

func (s *gitService) getContents(env, branch, subPath string) ([]byte, error) {
	path := s.resolvePath(env, subPath)
	contents, err := git.Contents(branch, path)
	if err != nil {
		return nil, err
	}
	if contents == nil && branch != "" {
		// Fall back to the default branch
		return git.Contents("", path)
	}
	return contents, nil
}

func (s *gitService) listDirectory(env, branch, subPath string) ([]string, error) {
	path := s.resolvePath(env, subPath)
	files, err := git.List(branch, path)
	if err != nil {
		return nil, err
	}
	if files == nil && branch != "" {
		// Fall back to the default branch
		return git.List("", path)
	}
	return files, nil
}

// Deployments returns a list of deployments for the given environment
func (s *gitService) Deployments(env, branch string) ([]string, error) {
	directoryContent, err := s.listDirectory(env, branch, "deployments")
	if err != nil {
		return nil, err
	}
	var deployments []string
	for _, filePath := range directoryContent {
		parts := strings.Split(filePath, ".")
		if len(parts) != 2 || parts[1] != "yaml" {
			continue
		}
		deployments = append(deployments, parts[0])
	}
	return deployments, nil
}

// Deployment returns a deployment for the given environment
func (s *gitService) Deployment(env, branch, name string) (Template, error) {
	fileContent, err := s.getContents(env, branch, "deployments/"+name+".yaml")
	if err != nil {
		return "", err
	}
	return Template(fileContent), nil
}

// Pods returns a list of pods for the given environment
func (s *gitService) Pods(env, branch string) ([]string, error) {
	directoryContent, err := s.listDirectory(env, branch, "pods")
	if err != nil {
		return nil, err
	}
	var pods []string
	for _, filePath := range directoryContent {
		parts := strings.Split(filePath, ".")
		if len(parts) != 2 || parts[1] != "yaml" {
			continue
		}
		pods = append(pods, parts[0])
	}
	return pods, nil
}

// Pod returns a list of pods for the given environment
func (s *gitService) Pod(env, branch, name string) (Template, error) {
	fileContent, err := s.getContents(env, branch, "pods/"+name+".yaml")
	if err != nil {
		return "", err
	}
	return Template(fileContent), nil
}

// Environment returns an environment template for the given branch
func (s *gitService) Environment(branch string) (Template, error) {
	fileContent, err := s.getContents("", branch, "environment.yaml")
	if err != nil {
		return "", err
	}
	return Template(fileContent), nil
}
