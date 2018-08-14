package api

import (
	"fmt"

	"github.com/airware/vili/circleci"
	"github.com/airware/vili/config"
	"github.com/airware/vili/environments"
	"github.com/airware/vili/kube"
	"github.com/airware/vili/log"
	"github.com/airware/vili/slack"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// InitializeCiClient initializes the CI client
func InitializeCiClient(ciProvider string) error {
	switch ciProvider {
	case "circleci":
		circleci.Init(&circleci.Config{
			Token:   config.GetString(config.CircleciToken),
			BaseURL: config.GetString(config.CircleciBaseurl),
		})
	default:
		log.WithField("CiProvider", ciProvider).Errorf("Unsupported ci Provider: %s", ciProvider)
	}
	return nil
}

// PostRolloutWebhook invokes a webhook on the provided ci
func PostRolloutWebhook(ciProvider string, env string) error {
	switch ciProvider {
	case "circleci":
		buildParameters := make(map[string]string)
		environment, err := environments.Get(env)
		namespace, err := kube.GetClient(environment.DeployedToEnv).Core().Namespaces().Get(environment.Name, metav1.GetOptions{})
		if err != nil {
			log.Warnf("Namespace with name %s doesn't exist", environment.Name)
			return nil
		}
		slack.PostLogMessage(fmt.Sprintf("Webhook invoked for *%s*", environment.Name), log.InfoLevel)
		buildParameters["CIRCLE_JOB"] = namespace.Annotations["vili.environment-webhook"]
		_, err = circleci.CircleBuild(config.GetString(config.GithubOwner), config.GetString(config.GithubRepo), environment.Branch, buildParameters)
		if err != nil {
			return err
		}
	default:
		log.Info("Define the post rollout webhook for you CI here")
	}
	return nil
}
