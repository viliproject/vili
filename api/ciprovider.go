package api

import (
	"fmt"

	"github.com/viliproject/vili/circleci"
	"github.com/viliproject/vili/config"
	"github.com/viliproject/vili/environments"
	"github.com/viliproject/vili/kube"
	"github.com/viliproject/vili/log"
	"github.com/viliproject/vili/slack"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// InitializeCiClient initializes the CI client
func InitializeCiClient(ciProvider string) error {
	switch ciProvider {
	case "circleci":
		err := circleci.Init(&circleci.Config{
			Token: config.GetString(config.CircleCIToken),
		})
		if err != nil {
			return err
		}
	case "":
		return nil
	default:
		return fmt.Errorf("Unsupported CI Provider: %s", ciProvider)
	}
	return nil
}

// PostRolloutWebhook invokes a webhook on the provided ci
func PostRolloutWebhook(ciProvider string, env string) error {
	switch ciProvider {
	case "circleci":
		environment, err := environments.Get(env)
		namespace, err := kube.GetClient(environment.DeployedToEnv).Core().Namespaces().Get(environment.Name, metav1.GetOptions{})
		if err != nil {
			log.Warnf("Namespace with name %s doesn't exist", environment.Name)
			return err
		}
		slack.PostLogMessage(fmt.Sprintf("Webhook invoked for *%s*", environment.Name), log.InfoLevel)
		buildParameters := map[string]string{
			"CIRCLE_JOB": namespace.Annotations["vili.environment-webhook"],
		}
		_, err = circleci.CircleBuild(config.GetString(config.GithubOwner), config.GetString(config.GithubRepo), environment.Branch, buildParameters)
		if err != nil {
			return err
		}
	default:
		log.Info("Define the post rollout webhook for you CI here")
	}
	return nil
}
