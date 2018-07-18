package repository

import (
	"context"
	"os"
	"testing"

	"github.com/airware/vili/log"
)

func TestRegistryGetRepository(t *testing.T) {
	ctx := context.Background()
	testService := &RegistryService{
		config: &RegistryConfig{
			Username: os.Getenv("REGISTRY_USERNAME"),
			Password: os.Getenv("REGISTRY_PASSWORD"),
		},
	}
	images, err := testService.GetRepository(ctx, "vili", []string{"master", "develop"})
	if err != nil {
		log.Error(err)
	}
	for _, image := range images {
		log.Info(image)
	}
}

func TestRegistryGetTag(t *testing.T) {
	ctx := context.Background()
	testService := &RegistryService{
		config: &RegistryConfig{
			Username: os.Getenv("REGISTRY_USERNAME"),
			Password: os.Getenv("REGISTRY_PASSWORD"),
		},
	}
	digest, err := testService.GetTag(ctx, "mysql", "master")
	if err != nil {
		log.Error(err)
	}
	log.Info(digest)
	digest, err = testService.GetTag(ctx, "quay.io/airware/vili", "1525471521-48b26ad")
	if err != nil {
		log.Error(err)
	}
	log.Info(digest)
}
