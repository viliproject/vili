package vili

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/viliproject/vili/api"
	"github.com/viliproject/vili/environments"
	"github.com/viliproject/vili/log"
	"github.com/viliproject/vili/redis"
	"github.com/viliproject/vili/repository"
	"github.com/viliproject/vili/slack"
	"github.com/viliproject/vili/util"
)

var publishedRegexp = regexp.MustCompile(
	`(?:[Pp]ublished) ([\w, ]+) with tag ([\w-]+) from branch ([\w/-]+)`,
)

// runDeployBot runs the deploy bot that listens to messages in the slack channel
func runDeployBot() {
	if !slack.IsAvailable() {
		return
	}
	mentionsRegexp, err := slack.MentionsRegexp()
	if err != nil {
		log.Error(err)
		return
	}
	mentions := make(chan *slack.Message)
	publishes := make(chan *slack.Message)
	go slack.ListenForMessages(map[*regexp.Regexp]chan<- *slack.Message{
		mentionsRegexp:  mentions,
		publishedRegexp: publishes,
	})

	for {
		select {
		case mention, ok := <-mentions:
			if !ok {
				mentions = nil
			} else {
				locked, err := redis.GetClient().SetNX(
					fmt.Sprintf("deploybotlock:%s", mention.Timestamp),
					true,
					1*time.Hour,
				).Result()
				if err != nil {
					log.Error(err)
					continue
				}
				if !locked {
					continue
				}
				err = handleCommand(strings.Fields(mention.Matches[0]), mention.Username)
				if err != nil {
					log.Error(err)
				}
			}

		case publish, ok := <-publishes:
			if !ok {
				publishes = nil
			} else {
				locked, err := redis.GetClient().SetNX(
					fmt.Sprintf("deploybotlock:%s", publish.Timestamp),
					true,
					1*time.Hour,
				).Result()
				if err != nil {
					log.Error(err)
					continue
				}
				if !locked {
					continue
				}
				images := strings.FieldsFunc(publish.Matches[0], func(c rune) bool {
					return !unicode.IsLetter(c) && !unicode.IsDigit(c)
				})
				tag := publish.Matches[1]
				branch := publish.Matches[2]
				err = handlePublish(images, tag, branch, publish.Username)
				if err != nil {
					log.Error(err)
				}
			}
		}

		if mentions == nil && publishes == nil {
			break
		}
	}
}

func handleCommand(command []string, username string) error {
	if len(command) == 0 {
		log.Debug("Skipping empty command")
		return nil
	}
	switch command[0] {
	case "deploy":
		if len(command) < 4 || len(command) > 5 {
			log.Debugf("Skipping invalid command %s", command)
			return nil
		}
		deployment := command[1]
		branch := command[2]
		tag := command[3]

		var env string

		if len(command) == 5 {
			env = command[4]
			if _, err := environments.Get(env); err != nil {
				log.Debugf("Invalid environment %s", env)
				return nil
			}
		} else {
			for _, e := range environments.Environments() {
				if e.Branch == branch && util.NewStringSet(e.Deployments).Contains(deployment) {
					env = e.Name
					break
				}
			}
			if env == "" {
				log.Debugf("No environment found for branch %s with deployment %s", branch, deployment)
				return nil
			}
		}
		rolloutDeployment(env, deployment, tag, branch, username)
	default:
		// TODO print usage?
		log.Debugf("Ignoring unknown command", command[0])
	}
	return nil
}

func handlePublish(images []string, tag, branch, username string) error {
	if len(images) == 0 {
		log.Debug("Skipping empty publish")
		return nil
	}

	for _, image := range images {
		var env string
		for _, e := range environments.Environments() {
			if e.Branch == branch {
				env = e.Name
				break
			}
		}
		if env != "" {
			rolloutDeployment(env, image, tag, branch, username)
		} else {
			log.Debugf("No deployment found for branch %s with name %s", branch, image)
		}
	}
	return nil
}

func rolloutDeployment(env, deployment, tag, branch, username string) {
	log.Debugf("Rolling out deployment %s, tag %s to env %s, requested by %s", deployment, tag, env, username)
	rollout := &api.Rollout{
		Env:            env,
		DeploymentName: deployment,
		Username:       username,
		Branch:         branch,
		Tag:            tag,
	}
	err := rollout.Run(true)
	if err != nil {
		switch e := err.(type) {
		case api.RolloutInitError:
			slack.PostLogMessage(e.Error(), log.ErrorLevel)
		case *repository.NotFoundError:
			slack.PostLogMessage(fmt.Sprintf("Deployment *%s* with tag *%s* not found", deployment, tag), log.ErrorLevel)
		default:
			log.Error(e)
		}
	}
}
