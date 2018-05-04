package repository

import (
	"os"
	"testing"

	"github.com/airware/vili/log"
	"github.com/stretchr/testify/assert"
)

func TestRegistryGetRepository(t *testing.T) {
	testService := &RegistryService{
		config: &RegistryConfig{
			BaseURL:   os.Getenv("REGISTRY_URL"),
			Username:  os.Getenv("REGISTRY_USERNAME"),
			Password:  os.Getenv("REGISTRY_PASSWORD"),
			Namespace: os.Getenv("REGISTRY_NAMESPACE"),
		},
	}
	images, err := testService.GetRepository("vili", []string{"master", "develop"})
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
			BaseURL:   os.Getenv("REGISTRY_URL"),
			Username:  os.Getenv("REGISTRY_USERNAME"),
			Password:  os.Getenv("REGISTRY_PASSWORD"),
			Namespace: os.Getenv("REGISTRY_NAMESPACE"),
		},
	}
	digest, err := testService.GetTag("vili", "master")
	if err != nil {
		log.Error(err)
	}
	log.Info(digest)
}

func TestRegistryFullName(t *testing.T) {
	for _, testCase := range []struct {
		RegistryConfig
		repo     string
		branch   string
		tag      string
		fullName string
	}{
		{
			RegistryConfig{
				BaseURL: "registry-1.docker.io",
			},
			"redis",
			"master",
			"1.9.1",
			"registry-1.docker.io/redis:master-1.9.1",
		},
		{
			RegistryConfig{
				BaseURL:   "quay.io",
				Namespace: "airware",
			},
			"vili",
			"testbranch",
			"abcdef",
			"quay.io/airware/vili:testbranch-abcdef",
		},
	} {
		testService := &RegistryService{&testCase.RegistryConfig}
		fullName, err := testService.FullName(testCase.repo, testCase.branch+"-"+testCase.tag)
		assert.NoError(t, err)
		assert.Equal(t, testCase.fullName, fullName)
	}
}
