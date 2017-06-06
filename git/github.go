package git

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// GithubConfig is the configuration for the github client
type GithubConfig struct {
	Token         string
	Owner         string
	Repo          string
	DefaultBranch string
}

type githubService struct {
	config *GithubConfig
	client *github.Client
}

// InitGithub initializes the github git service with the given config
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

func (s *githubService) getContents(branch, path string) (*github.RepositoryContent, []*github.RepositoryContent, *github.Response, error) {
	if branch == "" {
		branch = s.config.DefaultBranch
	}
	opts := &github.RepositoryContentGetOptions{Ref: branch}
	return s.client.Repositories.GetContents(context.TODO(), s.config.Owner, s.config.Repo, path, opts)
}

// Branches returns a list of branches for the repository
func (s *githubService) Branches() ([]string, error) {
	var opts *github.ListOptions
	var ret []string
	for {
		branches, resp, err := s.client.Repositories.ListBranches(context.TODO(), s.config.Owner, s.config.Repo, opts)
		if err != nil {
			return nil, err
		}
		for _, branch := range branches {
			ret = append(ret, *branch.Name)
		}
		if resp.NextPage > 0 {
			opts = &github.ListOptions{
				Page: resp.NextPage,
			}
		} else {
			return ret, nil
		}
	}
}

// Contents returns the contents of the file at the given path
func (s *githubService) Contents(branch, path string) (string, error) {
	file, _, _, err := s.getContents(branch, path)
	if err != nil {
		if errResp, ok := err.(*github.ErrorResponse); ok {
			if errResp.Response.StatusCode == 404 {
				return "", nil
			}
		}
		return "", err
	}
	if file == nil {
		return "", fmt.Errorf("%s is a directory", path)
	}
	return file.GetContent()
}

// List returns a list of subpaths of the given directory path
func (s *githubService) List(branch, path string) ([]string, error) {
	_, directory, _, err := s.getContents(branch, path)
	if err != nil {
		if errResp, ok := err.(*github.ErrorResponse); ok {
			if errResp.Response.StatusCode == 404 {
				return nil, nil
			}
		}
		return nil, err
	}
	if directory == nil {
		return nil, fmt.Errorf("%s is a file", path)
	}
	var paths []string
	prefix := strings.Split(path, "?")[0]
	for _, file := range directory {
		paths = append(paths, strings.TrimPrefix(strings.TrimPrefix(*file.Path, prefix), "/"))
	}
	return paths, nil
}
