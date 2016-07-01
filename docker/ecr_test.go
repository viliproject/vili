package docker

import (
	"os"
	"testing"

	"github.com/airware/vili/log"
)

func TestECRGetRepository(t *testing.T) {
	testService := newECR(&ECRConfig{
		Region:          os.Getenv("AWS_REGION"),
		AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		Namespace:       os.Getenv("REGISTRY_NAMESPACE"),
		BranchDelimiter: os.Getenv("REGISTRY_BRANCH_DELIMITER"),
	})
	images, err := testService.GetRepository("vili", true)
	if err != nil {
		log.Error(err)
	}
	for _, image := range images {
		log.Info(image)
	}
}

func TestECRGetTag(t *testing.T) {
	testService := newECR(&ECRConfig{
		Region:          os.Getenv("AWS_REGION"),
		AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		Namespace:       os.Getenv("REGISTRY_NAMESPACE"),
		BranchDelimiter: os.Getenv("REGISTRY_BRANCH_DELIMITER"),
	})
	digest, err := testService.GetTag("vili", "master", "latest")
	if err != nil {
		log.Error(err)
	}
	log.Info(digest)
}
