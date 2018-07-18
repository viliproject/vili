package repository

import (
	"context"
	"os"
	"testing"

	"github.com/airware/vili/log"
)

func TestECRGetRepository(t *testing.T) {
	ctx := context.Background()
	testService := newECR(&ECRConfig{
		Region:          os.Getenv("AWS_REGION"),
		AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
	})
	images, err := testService.GetRepository(ctx, "123", "vili", []string{"master", "develop"})
	if err != nil {
		log.Error(err)
	}
	for _, image := range images {
		log.Info(image)
	}
}

func TestECRGetTag(t *testing.T) {
	ctx := context.Background()
	testService := newECR(&ECRConfig{
		Region:          os.Getenv("AWS_REGION"),
		AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
	})
	digest, err := testService.GetTag(ctx, "123", "vili", "latest")
	if err != nil {
		log.Error(err)
	}
	log.Info(digest)
}
