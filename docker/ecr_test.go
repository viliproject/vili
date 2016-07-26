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
	images, err := testService.GetRepository("vili", []string{"master", "develop"})
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

func TestECRFullName(t *testing.T) {
	if os.Getenv("AWS_ACCOUNT_ID") == "" {
		t.Skip("No AWS_ACCOUNT_ID provided")
	}
	for _, testCase := range []struct {
		ECRConfig
		repo     string
		branch   string
		tag      string
		fullName string
	}{
		{
			ECRConfig{
				Region: os.Getenv("AWS_REGION"),
			},
			"baldr",
			"master",
			"latest",
			os.Getenv("AWS_ACCOUNT_ID") + ".dkr.ecr." + os.Getenv("AWS_REGION") + ".amazonaws.com/baldr:latest",
		},
		{
			ECRConfig{
				Region:          os.Getenv("AWS_REGION"),
				BranchDelimiter: "/",
			},
			"busybox",
			"feature/cld-9999",
			"latest",
			os.Getenv("AWS_ACCOUNT_ID") + ".dkr.ecr." + os.Getenv("AWS_REGION") + ".amazonaws.com/busybox/feature/cld-9999:latest",
		},
	} {
		testService := newECR(&testCase.ECRConfig)
		fullName, err := testService.FullName(testCase.repo, testCase.branch, testCase.tag)
		if err != nil {
			t.Error(err)
		} else if fullName != testCase.fullName {
			t.Errorf("%s != %s", fullName, testCase.fullName)
		}
	}
}
