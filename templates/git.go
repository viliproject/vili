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

func (s *gitService) getContents(env, branch, subPath string) (string, error) {
	path := s.resolvePath(env, subPath)
	contents, err := git.Contents(branch, path)
	if err != nil {
		return "", err
	}
	if contents == "" && branch != "" {
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

// Jobs returns a list of jobs for the given environment
func (s *gitService) Jobs(env, branch string) ([]string, error) {
	directoryContent, err := s.listDirectory(env, branch, "jobs")
	if err != nil {
		return nil, err
	}
	jobs := []string{}
	for _, filePath := range directoryContent {
		parts := strings.Split(filePath, ".")
		if len(parts) != 2 || parts[1] != "yaml" {
			continue
		}
		jobs = append(jobs, parts[0])
	}
	return jobs, nil
}

// Job returns a job for the given environment
func (s *gitService) Job(env, branch, name string) (Template, error) {
	fileContent, err := s.getContents(env, branch, "jobs/"+name+".yaml")
	if err != nil {
		return "", err
	}
	return Template(fileContent), nil
}

// Deployments returns a list of deployments for the given environment
func (s *gitService) Deployments(env, branch string) ([]string, error) {
	directoryContent, err := s.listDirectory(env, branch, "deployments")
	if err != nil {
		return nil, err
	}
	deployments := []string{}
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

// ConfigMaps returns a list of configMaps for the given environment
func (s *gitService) ConfigMaps(env, branch string) ([]string, error) {
	directoryContent, err := s.listDirectory(env, branch, "configmaps/"+env)
	if err != nil {
		return nil, err
	}
	configMaps := []string{}
	for _, filePath := range directoryContent {
		parts := strings.Split(filePath, ".")
		if len(parts) != 2 || parts[1] != "yaml" {
			continue
		}
		configMaps = append(configMaps, parts[0])
	}
	return configMaps, nil
}

// ConfigMap returns a configMap for the given environment
func (s *gitService) ConfigMap(env, branch, name string) (Template, error) {
	fileContent, err := s.getContents(env, branch, "configmaps/"+env+"/"+name+".yaml")
	if err != nil {
		return "", err
	}
	return Template(fileContent), nil
}

// Release returns a release template for the given environment
func (s *gitService) Release(env, branch string) (Template, error) {
	fileContent, err := s.getContents(env, branch, "release.yaml")
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
