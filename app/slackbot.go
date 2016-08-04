package vili

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/airware/vili/api"
	"github.com/airware/vili/docker"
	"github.com/airware/vili/environments"
	"github.com/airware/vili/log"
	"github.com/airware/vili/redis"
	"github.com/airware/vili/slack"
	"github.com/airware/vili/util"
)

var publishedRegexp = regexp.MustCompile(
	`(?:[Pp]ublish) ([\w, ]+) with tag ([\w-]+) from branch ([\w/-]+)`,
)

// runDeployBot runs the deploy bot that listens to messages in the slack channel
func runDeployBot() {
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
					return !unicode.IsLetter(c)
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
		app := command[1]
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
				if e.Branch == branch && util.NewStringSet(e.Apps).Contains(app) {
					env = e.Name
					break
				}
			}
			if env == "" {
				log.Debugf("No environment found for branch %s with app %s", branch, app)
				return nil
			}
		}
		deployApp(env, app, tag, branch, username)
	case "run":
		if len(command) < 4 || len(command) > 5 {
			log.Debugf("Skipping invalid command %s", command)
			return nil
		}
		job := command[1]
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
				if e.Branch == branch && util.NewStringSet(e.Jobs).Contains(job) {
					env = e.Name
					break
				}
			}
			if env == "" {
				log.Debugf("No environment found for branch %s with job %s", branch, job)
				return nil
			}
		}
		runJob(env, job, tag, branch, username)
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
		var isApp, isJob bool
		var env string
		for _, e := range environments.Environments() {
			if e.Branch == branch {
				if util.NewStringSet(e.Apps).Contains(image) {
					isApp = true
					env = e.Name
					break
				} else if util.NewStringSet(e.Jobs).Contains(image) {
					isJob = true
					env = e.Name
					break
				}
			}
		}
		if isApp {
			deployApp(env, image, tag, branch, username)
		} else if isJob {
			runJob(env, image, tag, branch, username)
		} else {
			log.Debugf("No app or job found for branch %s with name %s", branch, image)
		}
	}
	return nil
}

func deployApp(env, app, tag, branch, username string) {
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
}

func runJob(env, job, tag, branch, username string) {
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
}
