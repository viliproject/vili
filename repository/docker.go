package repository

import (
	"context"
	"regexp"
)

var ecrRegexp = regexp.MustCompile(`^([^.]+).dkr.ecr.([^.]+).amazonaws.com/(.+)$`)

// GetDockerRepository returns the images in the given repository for the provided branch names
func GetDockerRepository(ctx context.Context, repo string, branches []string) ([]*Image, error) {
	ecrMatch := ecrRegexp.FindStringSubmatch(repo)
	if len(ecrMatch) > 0 {
		return ecrService.GetRepository(ctx, ecrMatch[1], ecrMatch[3], branches)
	}
	return registryService.GetRepository(ctx, repo, branches)
}

// GetDockerTag returns an image digest for the given tag
func GetDockerTag(ctx context.Context, repo, tag string) (string, error) {
	ecrMatch := ecrRegexp.FindStringSubmatch(repo)
	if len(ecrMatch) > 0 {
		return ecrService.GetTag(ctx, ecrMatch[1], ecrMatch[3], tag)
	}
	return registryService.GetTag(ctx, repo, tag)
}
