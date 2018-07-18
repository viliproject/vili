package repository

import (
	"context"
	"fmt"
	"net/url"
)

// GetBundleRepository returns the images in the given repository for the provided branch names
func GetBundleRepository(ctx context.Context, repo string, branches []string) ([]*Image, error) {
	repoURL, err := url.Parse(repo)
	if err != nil {
		return nil, err
	}
	if repoURL.Scheme == "s3" {
		return s3Service.GetRepository(ctx, repoURL.Host, repoURL.Path, branches)
	}
	return nil, fmt.Errorf("Unknown bundle scheme: %s", repoURL.Scheme)
}
