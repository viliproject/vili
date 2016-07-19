package vili

import (
	"fmt"
	"strings"
	"time"

	"github.com/airware/vili/api"
	"github.com/airware/vili/docker"
	"github.com/airware/vili/log"
	"github.com/airware/vili/redis"
	"github.com/airware/vili/slack"
	"github.com/airware/vili/util"
)

// runDeployBot runs the deploy bot that listens to messages in the slack channel
func runDeployBot(envs *util.StringSet) {
	mentions := make(chan *slack.Mention)
	go slack.ListenForMentions(mentions)

	for mention := range mentions {
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
		err = handleCommand(strings.Fields(mention.Text), mention.Username, envs)
		if err != nil {
			log.Error(err)
		}
	}
}

func handleCommand(command []string, username string, envs *util.StringSet) error {
	if len(command) == 0 {
		log.Debug("Skipping empty command")
		return nil
	}
	switch command[0] {
	case "deploy":
		if len(command) != 5 {
			log.Debugf("Skipping invalid command %s", command)
			return nil
		}
		app := command[1]
		branch := command[2]
		tag := command[3]
		env := command[4]

		if !envs.Contains(env) {
			slack.PostLogMessage(fmt.Sprintf("Invalid environment *%s*", env), "error")
			return nil
		}

		log.Debugf("Deploying app %s, tag %s to env %s, requested by %s", app, tag, env, username)
		deployment := api.Deployment{
			Branch: branch,
			Tag:    tag,
		}
		err := deployment.Init(env, app, username, true)
		if err != nil {
			switch e := err.(type) {
			case api.DeploymentInitError:
				slack.PostLogMessage(e.Error(), "error")
			case *docker.NotFoundError:
				slack.PostLogMessage(fmt.Sprintf("App *%s* with tag *%s* not found", app, tag), "error")
			default:
				log.Error(e)
			}
		}
	case "run":
		if len(command) != 5 {
			log.Debugf("Skipping invalid command %s", command)
			return nil
		}
		job := command[1]
		branch := command[2]
		tag := command[3]
		env := command[4]

		if !envs.Contains(env) {
			slack.PostLogMessage(fmt.Sprintf("Invalid environment *%s*", env), "error")
			return nil
		}

		log.Debugf("Running job %s, tag %s in env %s, requested by %s", job, tag, env, username)
		run := api.Run{
			Branch: branch,
			Tag:    tag,
		}
		err := run.Init(env, job, username, true)
		if err != nil {
			switch e := err.(type) {
			case api.RunInitError:
				slack.PostLogMessage(e.Error(), "error")
			case *docker.NotFoundError:
				slack.PostLogMessage(fmt.Sprintf("Job *%s* with tag *%s* not found", job, tag), "error")
			default:
				log.Error(e)
			}
		}
	default:
		// TODO print usage?
		log.Debugf("Ignoring unknown command", command[0])
	}
	return nil
}
