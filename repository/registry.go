package repository

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/docker/distribution"
	"github.com/docker/distribution/reference"
	"github.com/docker/distribution/registry/client"
	"github.com/docker/distribution/registry/client/auth"
	"github.com/docker/distribution/registry/client/auth/challenge"
	"github.com/docker/distribution/registry/client/transport"
)

var registryService *RegistryService

// RegistryConfig is the registry service configuration
type RegistryConfig struct {
	Username string
	Password string
}

// RegistryService is an implementation of the docker Service interface
// It fetches docker images
type RegistryService struct {
	config *RegistryConfig
}

// InitRegistry initializes the docker registry service
func InitRegistry(c *RegistryConfig) error {
	registryService = &RegistryService{
		config: c,
	}
	return nil
}

// GetRepository implements the Service interface
func (s *RegistryService) GetRepository(ctx context.Context, repo string, branches []string) ([]*Image, error) {
	images, err := s.getImagesForBranches(ctx, repo, branches)
	if err != nil {
		return nil, err
	}

	sortByLastModified(images)
	return images, nil
}

// GetTag implements the Service interface
func (s *RegistryService) GetTag(ctx context.Context, repo, tag string) (string, error) {
	repository, err := s.getRepository(ctx, repo)
	if err != nil {
		return "", err
	}

	desc, err := repository.Tags(ctx).Get(ctx, tag)
	if err != nil {
		return "", err
	}

	return desc.Digest.String(), nil
}

func (s *RegistryService) getImagesForBranches(ctx context.Context, repoName string, branchNames []string) ([]*Image, error) {
	repo, err := s.getRepository(ctx, repoName)
	if err != nil {
		return nil, err
	}

	tags, err := repo.Tags(ctx).All(ctx)
	if err != nil {
		return nil, err
	}

	var images []*Image
	for _, tag := range tags {
		image := &Image{
			Tag: tag,
		}
		sepIndex := strings.LastIndex(tag, "-")
		if sepIndex != -1 {
			branchComponent, shaComponent := tag[:sepIndex], tag[sepIndex+1:]
			image.Revision = shaComponent
			for _, branchName := range branchNames {
				if branchComponent == slugFromBranch(branchName) {
					image.Branch = branchName
					images = append(images, image)
				}
			}
		}
	}
	return images, nil
}

func (s *RegistryService) getRepository(ctx context.Context, repoName string) (distribution.Repository, error) {
	repoNameRef, err := reference.ParseNormalizedNamed(repoName)
	if err != nil {
		return nil, err
	}
	domain := reference.Domain(repoNameRef)
	path := reference.Path(repoNameRef)

	baseURL := "https://" + domain
	if domain == "docker.io" {
		baseURL = "https://registry-1.docker.io"
	}

	credentialStore := &basicCredentialStore{
		Username: s.config.Username,
		Password: s.config.Password,
	}

	challengeManager := challenge.NewSimpleManager()
	resp, err := http.Get(baseURL + "/v2/")
	if err != nil {
		return nil, err
	}
	if err := challengeManager.AddResponse(resp); err != nil {
		return nil, err
	}

	transport := transport.NewTransport(http.DefaultTransport, auth.NewAuthorizer(
		challengeManager,
		auth.NewTokenHandler(http.DefaultTransport, credentialStore, path, "pull"),
		auth.NewBasicHandler(credentialStore),
	))

	repo, err := client.NewRepository(repoNameRef, baseURL, transport)
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
