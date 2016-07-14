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
			"registry-1.docker.io/redis:1.9.1",
		},
		{
			RegistryConfig{
				BaseURL:         "quay.io",
				Namespace:       "airware",
				BranchDelimiter: "-",
			},
			"vili",
			"testbranch",
			"abcdef",
			"quay.io/airware/vili-testbranch:abcdef",
		},
	} {
		testService := &RegistryService{&testCase.RegistryConfig}
		fullName, err := testService.FullName(testCase.repo, testCase.branch, testCase.tag)
		if err != nil {
			t.Error(err)
		} else if fullName != testCase.fullName {
			t.Errorf("%s != %s", fullName, testCase.fullName)
		}
	}
}
