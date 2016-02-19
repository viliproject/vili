package api

import (
	"fmt"
	"math"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/CloudCom/firego"
	"github.com/airware/vili/config"
	"github.com/airware/vili/errors"
	"github.com/airware/vili/kube"
	"github.com/airware/vili/kube/v1"
	"github.com/airware/vili/log"
	"github.com/airware/vili/redis"
	"github.com/airware/vili/slack"
	"github.com/airware/vili/templates"
)

const (
	rolloutPatchTimeout    = 10 * time.Second
	rolloutPatchPollPeriod = 1 * time.Second
	rolloutScaleTimeout    = 3 * time.Minute
	rolloutScalePollPeriod = 1 * time.Second
)

// deployer
type deployerSpec struct {
	env        string
	app        string
	deployment *Deployment
	db         *firego.Firebase

	// for all actions
	fromController *v1.ReplicationController
	toController   *v1.ReplicationController

	// for resume
	controllerTemplate templates.Template
	variables          map[string]string
	populatedTemplate  templates.Template
	lastPing           time.Time
}

func makeDeployer(env, app string, deployment *Deployment) (*deployerSpec, error) {
	deployer := &deployerSpec{
		env:        env,
		app:        app,
		deployment: deployment,
		db:         deploymentDB(env, app, deployment.ID),
		lastPing:   time.Now(),
	}
	controllerList, _, err := kube.Controllers.List(env, &url.Values{
		"labelSelector": []string{"app=" + app},
	})
	if err != nil {
		return nil, err
	}
	for _, controller := range controllerList.Items {
		if string(controller.ObjectMeta.UID) == deployment.FromUID {
			fromController := controller
			deployer.fromController = &fromController
		}
		if string(controller.ObjectMeta.UID) == deployment.ToUID {
			toController := controller
			deployer.toController = &toController
		}
	}

	if deployment.FromUID != "" && deployer.fromController == nil && deployer.toController == nil {
		deployer.db.Child("state").Set("completed")
		return nil, errors.BadRequestError("Could not find any controllers for this deployment")
	}

	return deployer, nil
}

func (d *deployerSpec) addMessage(message, level string) error {
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
	_, err := d.db.Child("log").Push(LogMessage{
		Time:    time.Now(),
		Message: message,
		Level:   level,
	})
	if err != nil {
		return err
	}

	if level != "debug" {
		urlStr := fmt.Sprintf(
			"%s/%s/apps/%s/deployments/%s",
			config.GetString(config.URI),
			d.env,
			d.app,
			d.deployment.ID,
		)
		slackMessage := fmt.Sprintf(
			"*%s* - *%s* - <%s|%s> - %s",
			d.env,
			d.app,
			urlStr,
			d.deployment.ID,
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

func (d *deployerSpec) resume() error {
	log.Infof("Resuming deployment %s for app %s in env %s", d.deployment.ID, d.app, d.env)
	var waitGroup sync.WaitGroup
	failed := false

	// controllerTemplate
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		body, err := templates.Controller(d.env, d.app)
		if err != nil {
			log.Error(err)
			failed = true
			return
		}
		d.controllerTemplate = body
	}()

	// variables
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		variables, err := templates.Variables(d.env)
		if err != nil {
			log.Error(err)
			failed = true
			return
		}
		d.variables = variables
	}()

	waitGroup.Wait()
	if failed {
		return fmt.Errorf("Failed one of the service calls")
	}

	// combine the template with the variables to get the populated template
	populatedTemplate, invalid := d.controllerTemplate.Populate(d.variables)
	if invalid {
		return errors.BadRequestError("Controller template missing variables")
	}
	d.populatedTemplate = populatedTemplate

	// return any lock errors synchronously
	if err := d.acquireLock(); err != nil {
		return err
	}

	// if no problem fetching lock, run the rollout asynchronously
	go func() {
		defer d.releaseLock()

		var state string
		d.db.Child("state").Value(&state)
		var message string
		switch state {
		case deploymentStateNew:
			message = fmt.Sprintf("Starting deployment created by *%s* for tag *%s*", d.deployment.Username, d.deployment.Tag)
		case deploymentStatePaused:
			message = "Resuming deployment"
		default:
			log.Errorf("Cannot resume deployment from state %s", state)
			return
		}
		d.db.Child("state").Set(deploymentStateRunning)
		d.deployment.State = deploymentStateRunning
		d.addMessage(message, "info")

		// fetch the latest pods for the fromController
		if d.fromController != nil {
			if _, _, err := d.refreshControllerPodCount(d.fromController); err != nil {
				log.Error(err)
				return
			}
		}

		if d.toController == nil {
			toController, err := d.createNewController(d.app+"-"+d.deployment.ID, 0)
			if err != nil {
				log.Error(err)
				return
			}
			d.toController = toController
			d.db.Child("toUid").Set(toController.ObjectMeta.UID)
			d.deployment.ToUID = string(toController.ObjectMeta.UID)
			d.addMessage(fmt.Sprintf("Created new controller: %s", toController.ObjectMeta.Name), "debug")
		}

		stateNotifications := make(chan firego.Event)
		stateRef := d.db.Child("state")
		if err := stateRef.Watch(stateNotifications); err != nil {
			log.Error(err)
			return
		}
		defer stateRef.StopWatching()
		go func() {
			for event := range stateNotifications {
				log.Debugf("Deployment state changed to %s", event.Data.(string))
				d.deployment.State = event.Data.(string)
			}
		}()

		err := d.rollout()
		if err != nil {
			d.db.Child("state").Set(deploymentStatePaused)
			switch e := err.(type) {
			case deployerPaused:
				d.addMessage("Paused deployment", "warn")
			case deployerAutoPaused:
				// pass
			case *deployerTimeout:
				d.addMessage(e.message, "error")
			default:
				log.Error(err)
			}
		}
	}()
	return nil
}

func (d *deployerSpec) pause() error {
	log.Infof("Pausing deployment %s for app %s in env %s", d.deployment.ID, d.app, d.env)
	var state string
	var newState string
	d.db.Child("state").Value(&state)
	var message string
	switch state {
	case deploymentStateRunning:
		message = "Pausing deployment"
		newState = deploymentStatePausing
	case deploymentStatePausing:
		message = "Force pausing deployment"
		newState = deploymentStatePaused
	default:
		return errors.BadRequestError(fmt.Sprintf("Cannot pause deployment from state %s", state))
	}
	d.db.Child("state").Set(newState)
	d.addMessage(message, "warn")
	return nil
}

func (d *deployerSpec) rollback() error {
	log.Infof("Rolling back deployment %s for app %s in env %s", d.deployment.ID, d.app, d.env)
	if err := d.acquireLock(); err != nil {
		return err
	}
	defer d.releaseLock()

	var state string
	var newState string
	d.db.Child("state").Value(&state)
	var message string
	switch state {
	case deploymentStatePaused:
		message = "Rolling back deployment"
		newState = deploymentStateRollingback
	default:
		return errors.BadRequestError(fmt.Sprintf("Cannot roll back deployment from state %s", state))
	}
	d.db.Child("state").Set(newState)
	d.deployment.State = newState
	d.addMessage(message, "warn")

	if d.fromController != nil && d.toController != nil {
		if err := d.scaleController(d.fromController, d.deployment.DesiredReplicas); err != nil {
			return err
		}
		if err := d.scaleController(d.toController, 0); err != nil {
			return err
		}
		if _, _, err := kube.Controllers.Delete(d.env, d.toController.ObjectMeta.Name); err != nil {
			return err
		}
	} else {
		log.Warn("Cannot roll back deployment with no from or to controller")
	}
	d.db.Child("state").Set(deploymentStateRolledback)
	d.addMessage("Rolled back deployment", "warn")
	return nil
}

// utils
func (d *deployerSpec) rollout() error {
	replicas, _, err := d.refreshControllerPodCount(d.toController)
	if err != nil {
		return err
	}
	nextSteps := d.getNextStepsForCurrentCount(replicas)
	log.Debugf("Next steps in rollout are %s", nextSteps)
	for _, desiredToReplicas := range nextSteps {
		desiredFromReplicas := d.deployment.DesiredReplicas - desiredToReplicas
		// ping
		if err := d.ping(); err != nil {
			return err
		}
		// scale to controller
		if err := d.scaleController(d.toController, desiredToReplicas); err != nil {
			return err
		}
		// scale from controller
		if d.fromController != nil {
			fromReplicas, _, err := d.refreshControllerPodCount(d.fromController)
			if err != nil {
				return err
			}
			if fromReplicas > desiredFromReplicas {
				if err := d.scaleController(d.fromController, desiredFromReplicas); err != nil {
					return err
				}
			}
		}

		if len(nextSteps) > 1 && d.deployment.Rollout != nil && d.deployment.Rollout.Autopause {
			return deployerAutoPaused{}
		}
	}

	// rename
	// first, delete the from controller, deleting it's pods if necessary
	if d.fromController != nil {
		fromReplicas, _, err := d.refreshControllerPodCount(d.fromController)
		if err != nil {
			return err
		}
		if fromReplicas > 0 {
			_, status, err := kube.Pods.DeleteForController(d.env, d.fromController)
			if err != nil {
				return err
			}
			if status != nil {
				return fmt.Errorf("failed deleting pods for controller")
			}
			_, _, err = d.refreshControllerPodCount(d.fromController)
			if err != nil {
				return err
			}
			d.addMessage(fmt.Sprintf("Scaled %s to 0 replicas", d.fromController.ObjectMeta.Name), "debug")
		}
		if _, _, err := kube.Controllers.Delete(d.env, d.fromController.ObjectMeta.Name); err != nil {
			return err
		}
	}
	// then create a new controller by copying the to controller
	_, err = d.createNewController(d.app, d.deployment.DesiredReplicas)
	if err != nil {
		return err
	}

	// then delete the to controller
	if _, _, err := kube.Controllers.Delete(d.env, d.toController.ObjectMeta.Name); err != nil {
		return err
	}
	d.addMessage(fmt.Sprintf("Renamed %s to %s", d.toController.ObjectMeta.Name, d.app), "debug")

	// ping to update the clock
	if err := d.ping(); err != nil {
		return err
	}

	d.addMessage(fmt.Sprintf("Successfully completed rollout in %s", d.deployment.Clock.humanize()), "info")
	d.db.Child("state").Set(deploymentStateCompleted)

	return nil
}

func (d *deployerSpec) scaleController(controller *v1.ReplicationController, desiredReplicas int) error {
	if err := d.ping(); err != nil {
		return err
	}

	// try to set the number of replicas
	elapsed := 0 * time.Second
	patchTicker := time.NewTicker(rolloutPatchPollPeriod)
	defer patchTicker.Stop()
	for {
		_, status, err := kube.Controllers.Patch(d.env, controller.ObjectMeta.Name, &v1.ReplicationController{
			Spec: v1.ReplicationControllerSpec{
				Replicas: &desiredReplicas,
			},
		})
		if err != nil {
			return err
		}
		if status == nil {
			break
		}
		log.Warn("Failed scaling attempt, retrying")
		// tick
		<-patchTicker.C
		elapsed += rolloutPatchPollPeriod
		if elapsed > rolloutPatchTimeout {
			return fmt.Errorf("Failed scaling controller %s", controller.ObjectMeta.Name)
		}
	}
	d.addMessage(fmt.Sprintf("Scaling %s to %d replicas", controller.ObjectMeta.Name, desiredReplicas), "debug")

	// wait for scale
	elapsed = 0 * time.Second
	waitTicker := time.NewTicker(rolloutScalePollPeriod)
	defer waitTicker.Stop()
	for {
		if err := d.ping(); err != nil {
			return err
		}
		readyCount, runningCount, err := d.refreshControllerPodCount(controller)
		if err != nil {
			return err
		}
		if readyCount == runningCount && readyCount == desiredReplicas {
			break
		}
		// tick
		<-waitTicker.C
		elapsed += rolloutScalePollPeriod
		if elapsed > rolloutScaleTimeout {
			return &deployerTimeout{
				message: fmt.Sprintf(
					"Timed out waiting for controller %s to scale",
					controller.ObjectMeta.Name,
				),
			}
		}
	}
	d.addMessage(fmt.Sprintf("Scaled %s to %d replicas", controller.ObjectMeta.Name, desiredReplicas), "debug")
	return nil
}

func (d *deployerSpec) acquireLock() error {
	locked, err := redis.GetClient().SetNX(
		fmt.Sprintf("deploymentlock:%s:%s", d.env, d.app),
		true,
		1*time.Hour,
	).Result()
	if err != nil {
		return err
	}
	if !locked {
		return errors.ConflictError("Failed to acquire deployment lock")
	}
	log.Debugf("Acquired lock for %s in %s", d.app, d.env)
	WaitGroup.Add(1)
	return nil
}

func (d *deployerSpec) releaseLock() {
	WaitGroup.Done()
	err := redis.GetClient().Del(
		fmt.Sprintf("deploymentlock:%s:%s", d.env, d.app),
	).Err()
	if err != nil {
		log.Error(err)
	} else {
		log.Debugf("Released lock for %s in %s", d.app, d.env)
	}
}

// ping updates the deployment clock, and checks if the deployment should be paused
func (d *deployerSpec) ping() error {
	now := time.Now()
	elapsed := Clock(now.Sub(d.lastPing))
	if d.deployment.Clock == nil {
		d.deployment.Clock = &elapsed
	} else {
		*d.deployment.Clock += elapsed
	}
	d.lastPing = now
	d.db.Child("clock").Set(d.deployment.Clock)

	if d.deployment.State == deploymentStatePausing ||
		d.deployment.State == deploymentStatePaused {
		return deployerPaused{}
	}
	if Exiting {
		log.Warn("Pausing deployment after server shutdown request")
		return deployerPaused{}
	}
	return nil
}

func (d *deployerSpec) refreshControllerPodCount(controller *v1.ReplicationController) (int, int, error) {
	kubePods, status, err := kube.Pods.ListForController(d.env, controller)
	if err != nil {
		return 0, 0, err
	}
	if status != nil {
		return 0, 0, fmt.Errorf("Could not fetch the list of pods for controller %s", controller.ObjectMeta.Name)
	}
	pods, readyCount, runningCount := getPodsFromPodList(kubePods)

	if string(controller.ObjectMeta.UID) == d.deployment.FromUID {
		d.db.Child("fromPods").Set(pods)
	} else if string(controller.ObjectMeta.UID) == d.deployment.ToUID {
		d.db.Child("toPods").Set(pods)
	}

	return readyCount, runningCount, nil
}

func (d *deployerSpec) createNewController(name string, replicas int) (*v1.ReplicationController, error) {
	controller := &v1.ReplicationController{}
	err := d.populatedTemplate.Parse(controller)
	if err != nil {
		return nil, err
	}

	containers := controller.Spec.Template.Spec.Containers
	if len(containers) == 0 {
		return nil, fmt.Errorf("no containers in controller")
	}
	image := containers[0].Image
	imageSplit := strings.Split(image, ":")
	if len(imageSplit) != 2 {
		return nil, fmt.Errorf("invalid image: %s", image)
	}
	imageSplit[1] = d.deployment.Tag
	containers[0].Image = strings.Join(imageSplit, ":")

	controller.ObjectMeta.Name = name
	controller.Spec.Replicas = &replicas
	controller.Spec.Selector["deployment"] = d.deployment.ID
	controller.Spec.Template.ObjectMeta.Labels["deployment"] = d.deployment.ID

	resp, status, err := kube.Controllers.Create(d.env, controller)
	if status != nil {
		return nil, fmt.Errorf(status.Message)
	}
	return resp, err
}

func newFloat64(f float64) *float64 {
	return &f
}

var defaultRolloutStrategy = &RolloutStrategy{
	Name: "33% - 67% - 100%",
	Steps: []RolloutStrategyStep{
		RolloutStrategyStep{Ratio: newFloat64(0.33)},
		RolloutStrategyStep{Ratio: newFloat64(0.67)},
		RolloutStrategyStep{Ratio: newFloat64(1)},
	},
}

func (d *deployerSpec) getNextStepsForCurrentCount(count int) []int {
	var nextSteps []int
	if count < d.deployment.DesiredReplicas {
		strategy := defaultRolloutStrategy
		if d.deployment.Rollout != nil && d.deployment.Rollout.Strategy != nil {
			strategy = d.deployment.Rollout.Strategy
		}

		prevCount := count
		for _, step := range strategy.Steps {
			var stepCount int
			if step.Count != nil {
				stepCount = *step.Count
			} else if step.Ratio != nil && *step.Ratio <= 1 {
				stepCount = int(math.Floor((float64(d.deployment.DesiredReplicas) * *step.Ratio) + 0.5))
			} else {
				continue
			}

			if stepCount > prevCount {
				nextSteps = append(nextSteps, stepCount)
				prevCount = stepCount
			}
		}
		if prevCount != d.deployment.DesiredReplicas {
			nextSteps = append(nextSteps, d.deployment.DesiredReplicas)
		}
	}
	return nextSteps
}

type deployerException struct {
}

func (e *deployerException) Error() string {
	return "Deployer exception"
}

type deployerPaused struct {
	*deployerException
}

type deployerAutoPaused struct {
	*deployerException
}

type deployerTimeout struct {
	message string
}

func (e *deployerTimeout) Error() string {
	return e.message
}
