package docker

import (
	"os"
	"testing"

	"github.com/airware/vili/log"
)

func TestQuayGetRepository(t *testing.T) {
	testService := &QuayService{
		config: &QuayConfig{
			Token:     os.Getenv("QUAY_TOKEN"),
			Namespace: os.Getenv("QUAY_NAMESPACE"),
		},
	}
	images, err := testService.GetRepository("skadi", true)
	if err != nil {
		log.Error(err)
	}
	for _, image := range images {
		log.Info(image)
	}
}
