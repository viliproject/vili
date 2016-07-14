package api

import (
	"fmt"
	"sync"
	"time"

	"github.com/CloudCom/firego"
	"github.com/airware/vili/config"
	"github.com/airware/vili/docker"
	"github.com/airware/vili/errors"
	"github.com/airware/vili/kube"
	"github.com/airware/vili/kube/v1"
	"github.com/airware/vili/log"
	"github.com/airware/vili/redis"
	"github.com/airware/vili/slack"
	"github.com/airware/vili/templates"
)

const (
	runTimeout    = 5 * time.Minute
	runPollPeriod = 1 * time.Second
)

// runner
type runnerSpec struct {
	env string
	job string
	run *Run
	db  *firego.Firebase

	// for all actions
	pod *v1.Pod

	// for resume
	podTemplate       templates.Template
	variables         map[string]string
	populatedTemplate templates.Template
	lastPing          time.Time
}

func makeRunner(env, job string, run *Run) (*runnerSpec, error) {
	runner := &runnerSpec{
		env:      env,
		job:      job,
		run:      run,
		db:       runDB(env, job, run.ID),
		lastPing: time.Now(),
	}
	return runner, nil
}

func (r *runnerSpec) addMessage(message, level string) error {
	var logf func(...interface{})
	switch level {
	case "debug":
		logf = log.Debug
	case "info":
		logf = log.Info
	case "warn":
		logf = log.Warn
	case "error":
		logf = log.Error
	default:
		return fmt.Errorf("Invalid level %s", level)
	}
	logf(message)
	_, err := r.db.Child("log").Push(LogMessage{
		Time:    time.Now(),
		Message: message,
		Level:   level,
	})
	if err != nil {
		return err
	}

	if level != "debug" {
		urlStr := fmt.Sprintf(
			"%s/%s/jobs/%s/runs/%s",
			config.GetString(config.URI),
			r.env,
			r.job,
			r.run.ID,
		)
		slackMessage := fmt.Sprintf(
			"*%s* - *%s* - <%s|%s> - %s",
			r.env,
			r.job,
			urlStr,
			r.run.ID,
			message,
		)
		if level == "error" {
			slackMessage += " <!channel>"
		}
		err := slack.PostLogMessage(slackMessage, level)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *runnerSpec) start() error {
	log.Infof("Starting run %s for job %s in env %s", r.run.ID, r.job, r.env)
	var waitGroup sync.WaitGroup
	failed := false

	// podTemplate
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		body, err := templates.Pod(r.env, r.job)
		if err != nil {
			log.Error(err)
			failed = true
			return
		}
		r.podTemplate = body
	}()

	// variables
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		variables, err := templates.Variables(r.env)
		if err != nil {
			log.Error(err)
			failed = true
			return
		}
		r.variables = variables
	}()

	waitGroup.Wait()
	if failed {
		return fmt.Errorf("Failed one of the service calls")
	}

	// combine the template with the variables to get the populated template
	populatedTemplate, invalid := r.podTemplate.Populate(r.variables)
	if invalid {
		return errors.BadRequestError("Pod template missing variables")
	}
	r.populatedTemplate = populatedTemplate

	err := r.acquireLock()
	// return any lock errors synchronously
	if err != nil {
		return err
	}

	// if no problem fetching lock, run the rollout asynchronously
	go func() {
		defer r.releaseLock()

		var state string
		r.db.Child("state").Value(&state)
		var message string
		switch state {
		case runStateNew:
			message = fmt.Sprintf("Starting job run created by *%s* for tag *%s*", r.run.Username, r.run.Tag)
		default:
			log.Errorf("Cannot resume run from state %s", state)
			return
		}
		r.db.Child("state").Set(runStateRunning)
		r.run.State = runStateRunning
		r.addMessage(message, "info")

		stateNotifications := make(chan firego.Event)
		stateRef := r.db.Child("state")
		if err := stateRef.Watch(stateNotifications); err != nil {
			log.Error(err)
			return
		}
		defer stateRef.StopWatching()
		go func() {
			for event := range stateNotifications {
				log.Debugf("Run state changed to %s", event.Data.(string))
				r.run.State = event.Data.(string)
			}
		}()

		// create pod, which starts the job
		pod, err := r.createNewPod()
		if err != nil {
			log.Error(err)
			return
		}
		r.pod = pod
		r.run.UID = string(pod.ObjectMeta.UID)
		r.db.Child("uid").Set(pod.ObjectMeta.UID)

		// wait for completion
		err = r.waitForPod()
		if err != nil {
			r.db.Child("state").Set(runStateTerminated)
			switch e := err.(type) {
			case runnerTerminated:
				r.addMessage("Terminated run", "warn")
			case runnerTimeout:
				r.addMessage("Run timed out", "error")
			default:
				r.addMessage(fmt.Sprintf("Unexpected error %s", e), "error")
				log.Error(e)
			}
		}
	}()
	return nil
}

func (r *runnerSpec) terminate() error {
	log.Infof("Terminating run %s for job %s in env %s", r.run.ID, r.job, r.env)
	var state string
	var newState string
	r.db.Child("state").Value(&state)
	var message string
	switch state {
	case runStateRunning:
		message = "Pausing run"
		newState = runStateTerminating
	case runStateTerminating:
		message = "Force terminating run"
		newState = runStateTerminated
	default:
		return errors.BadRequestError(fmt.Sprintf("Cannot terminate run from state %s", state))
	}
	r.addMessage(message, "warn")
	r.db.Child("state").Set(newState)
	return nil
}

// utils
func (r *runnerSpec) acquireLock() error {
	locked, err := redis.GetClient().SetNX(
		fmt.Sprintf("runlock:%s:%s", r.env, r.job),
		true,
		1*time.Hour,
	).Result()
	if err != nil {
		return err
	}
	if !locked {
		return errors.ConflictError("Failed to acquire run lock")
	}
	WaitGroup.Add(1)
	return nil
}

func (r *runnerSpec) releaseLock() error {
	WaitGroup.Done()
	return redis.GetClient().Del(
		fmt.Sprintf("runlock:%s:%s", r.env, r.job),
	).Err()
}

// waitForPod waits until the pod exits
func (r *runnerSpec) waitForPod() error {
	// try to set the number of replicas
	elapsed := 0 * time.Second
	ticker := time.NewTicker(runPollPeriod)
	defer ticker.Stop()
WaitLoop:
	for {
		if err := r.ping(); err != nil {
			return err
		}

		pod, _, err := kube.Pods.Get(r.env, r.pod.ObjectMeta.Name)
		if err != nil {
			return err
		}
		if pod != nil {
			log, _, err := kube.Pods.GetLog(r.env, r.pod.ObjectMeta.Name)
			if err != nil {
				return err
			}
			if log != "" {
				r.db.Child("output").Set(log)
			}
			switch pod.Status.Phase {
			case v1.PodSucceeded:
				r.db.Child("state").Set(runStateCompleted)
				_, _, err := kube.Pods.Delete(r.env, r.pod.ObjectMeta.Name)
				if err != nil {
					return err
				}
				r.addMessage(fmt.Sprintf("Successfully completed job in %s", r.run.Clock.humanize()), "info")
				break WaitLoop
			case v1.PodFailed:
				r.db.Child("state").Set(runStateFailed)
				r.db.Child("stateReason").Set("Pod failed")
				break WaitLoop
			}
		}

		// tick
		<-ticker.C
		elapsed += runPollPeriod
		if elapsed > runTimeout {
			return runnerTimeout{}
		}
	}
	return nil
}

// ping updates the run clock, and checks if the run should be terminated
func (r *runnerSpec) ping() error {
	now := time.Now()
	elapsed := Clock(now.Sub(r.lastPing))
	if r.run.Clock == nil {
		r.run.Clock = &elapsed
	} else {
		*r.run.Clock += elapsed
	}
	r.lastPing = now
	r.db.Child("clock").Set(r.run.Clock)

	if r.run.State == runStateTerminating ||
		r.run.State == runStateTerminated {
		return runnerTerminated{}
	}

	if Exiting {
		log.Warn("Terminating run after server shutdown request")
		return runnerTerminated{}
	}
	return nil
}

func (r *runnerSpec) createNewPod() (*v1.Pod, error) {
	pod := &v1.Pod{}
	err := r.populatedTemplate.Parse(pod)
	if err != nil {
		return nil, err
	}

	containers := pod.Spec.Containers
	if len(containers) == 0 {
		return nil, fmt.Errorf("no containers in pod")
	}

	imageName, err := docker.FullName(r.job, r.run.Branch, r.run.Tag)
	if err != nil {
		return nil, err
	}
	containers[0].Image = imageName

	pod.ObjectMeta.Name = r.job + "-" + r.run.ID
	pod.ObjectMeta.Labels = map[string]string{
		"job": r.job,
		"run": r.run.ID,
	}

	resp, status, err := kube.Pods.Create(r.env, pod)
	if status != nil {
		return nil, fmt.Errorf(status.Message)
	}
	return resp, err
}

type runnerException struct {
}

func (e *runnerException) Error() string {
	return "Runner exception"
}

type runnerTerminated struct {
	*runnerException
}

type runnerTimeout struct {
	*runnerException
}
