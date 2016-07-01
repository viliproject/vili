package docker

import (
	"os"
	"testing"

	"github.com/airware/vili/log"
)

func TestRegistryGetRepository(t *testing.T) {
	testService := &RegistryService{
		config: &RegistryConfig{
			BaseURL:         os.Getenv("REGISTRY_URL"),
			Username:        os.Getenv("REGISTRY_USERNAME"),
			Password:        os.Getenv("REGISTRY_PASSWORD"),
			Namespace:       os.Getenv("REGISTRY_NAMESPACE"),
			BranchDelimiter: os.Getenv("REGISTRY_BRANCH_DELIMITER"),
		},
	}
	images, err := testService.GetRepository("vili", true)
	if err != nil {
		log.Error(err)
	}
	for _, image := range images {
		log.Info(image)
	}
}

func TestRegistryGetTag(t *testing.T) {
	testService := &RegistryService{
		config: &RegistryConfig{
			BaseURL:         os.Getenv("REGISTRY_URL"),
			Username:        os.Getenv("REGISTRY_USERNAME"),
			Password:        os.Getenv("REGISTRY_PASSWORD"),
			Namespace:       os.Getenv("REGISTRY_NAMESPACE"),
			BranchDelimiter: os.Getenv("REGISTRY_BRANCH_DELIMITER"),
		},
	}
	digest, err := testService.GetTag("vili", "master", "master")
	if err != nil {
		log.Error(err)
	}
	log.Info(digest)
}
