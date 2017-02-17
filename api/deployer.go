package api

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/CloudCom/firego"
	"github.com/airware/vili/config"
	"github.com/airware/vili/docker"
	"github.com/airware/vili/errors"
	"github.com/airware/vili/kube"
	"github.com/airware/vili/kube/extensions/v1beta1"
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
	deployment *Deployment
	db         *firego.Firebase

	// for all actions
	fromReplicaSet *v1beta1.ReplicaSet
	toReplicaSet   *v1beta1.ReplicaSet

	// for resume
	deploymentTemplate templates.Template
	variables          map[string]string
	lastPing           time.Time
}

func makeDeployer(deployment *Deployment) (*deployerSpec, error) {
	deployer := &deployerSpec{
		deployment: deployment,
		db:         deploymentDB(deployment.Env, deployment.App, deployment.ID),
		lastPing:   time.Now(),
	}
	replicaSetList, _, err := kube.ReplicaSets.List(deployment.Env, &url.Values{
		"labelSelector": []string{"app=" + deployment.App},
	})
	if err != nil {
		return nil, err
	}

	if replicaSetList != nil {
		for _, replicaSet := range replicaSetList.Items {
			if string(replicaSet.ObjectMeta.UID) == deployment.FromUID {
				fromReplicaSet := replicaSet
				deployer.fromReplicaSet = &fromReplicaSet
			}
			if string(replicaSet.ObjectMeta.UID) == deployment.ToUID {
				toReplicaSet := replicaSet
				deployer.toReplicaSet = &toReplicaSet
			}
		}
	}

	if deployment.FromUID != "" && deployer.fromReplicaSet == nil && deployer.toReplicaSet == nil {
		log.Warn("Could not find any replica sets for this deployment")
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
			d.deployment.Env,
			d.deployment.App,
			d.deployment.ID,
		)
		slackMessage := fmt.Sprintf(
			"*%s* - *%s* - <%s|%s> - %s",
			d.deployment.Env,
			d.deployment.App,
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
	log.Infof("Resuming deployment %s for app %s in env %s", d.deployment.ID, d.deployment.App, d.deployment.Env)

	body, err := templates.Deployment(d.deployment.Env, d.deployment.Branch, d.deployment.App)
	if err != nil {
		return err
	}
	d.deploymentTemplate = body

	deployment := &v1beta1.Deployment{}
	err = d.deploymentTemplate.Parse(deployment)
	if err != nil {
		return err
	}

	deployment.Spec.Template.ObjectMeta.Labels["deployment"] = d.deployment.ID

	imageName, err := docker.FullName(d.deployment.App, d.deployment.Branch, d.deployment.Tag)
	if err != nil {
		return err
	}
	deployment.Spec.Template.Spec.Containers[0].Image = imageName

	if d.deployment.DesiredReplicas == 0 {
		if d.fromReplicaSet != nil && d.fromReplicaSet.Spec.Replicas != nil {
			d.deployment.DesiredReplicas = int(*d.fromReplicaSet.Spec.Replicas)
		} else {
			d.deployment.DesiredReplicas = int(*deployment.Spec.Replicas)
		}
		d.db.Child("desiredReplicas").Set(d.deployment.DesiredReplicas)
	} else {
		*deployment.Spec.Replicas = int32(d.deployment.DesiredReplicas)
	}

	if d.deployment.Rollout != nil {
		deployment.Spec.Strategy.Type = v1beta1.DeploymentStrategyType(d.deployment.Rollout.Strategy)
		if d.deployment.Rollout.Strategy != rolloutStrategyRollingUpdate {
			deployment.Spec.Strategy.RollingUpdate = nil
		}
	}

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
			message = fmt.Sprintf("Starting deployment created by *%s* for tag *%s* on branch *%s*", d.deployment.Username, d.deployment.Tag, d.deployment.Branch)
		case deploymentStatePaused:
			message = "Resuming deployment"
		default:
			log.Errorf("Cannot resume deployment from state %s", state)
			return
		}
		d.db.Child("state").Set(deploymentStateRunning)
		d.deployment.State = deploymentStateRunning
		d.addMessage(message, "info")

		// fetch the latest pods for the fromReplicaSet
		if d.fromReplicaSet != nil {
			if _, _, err := d.refreshReplicaSetPodCount(d.fromReplicaSet); err != nil {
				log.Error(err)
				return
			}
		}

		newDeployment, _, err := kube.Deployments.Replace(d.deployment.Env, d.deployment.App, deployment)
		if err != nil {
			log.Error(err)
			return
		}

		if newDeployment == nil {
			newDeployment, _, err = kube.Deployments.Create(d.deployment.Env, deployment)
			if err != nil {
				log.Error(err)
				return
			}
		}

		// It can take some time for kubernetes to populate the revision annotation
		for {
			if _, ok := newDeployment.ObjectMeta.Annotations["deployment.kubernetes.io/revision"]; ok {
				break
			}
			time.Sleep(100 * time.Millisecond)
			newDeployment, _, err = kube.Deployments.Get(d.deployment.Env, d.deployment.App)
			if err != nil {
				log.Error(err)
				return
			}
		}

		replicaSetList, _, err := kube.ReplicaSets.ListForDeployment(d.deployment.Env, newDeployment)
		if err != nil {
			log.Error(err)
			return
		}
		newRevision, ok := newDeployment.ObjectMeta.Annotations["deployment.kubernetes.io/revision"]
		if !ok {
			log.Error("No revision annotation found on the new deployment")
			return
		}
		var toReplicaSet *v1beta1.ReplicaSet
		if replicaSetList != nil {
			for _, replicaSet := range replicaSetList.Items {
				revision := replicaSet.ObjectMeta.Annotations["deployment.kubernetes.io/revision"]
				if revision == newRevision {
					toReplicaSet = &replicaSet
					break
				}
			}
		}
		if toReplicaSet == nil {
			log.Error("Unable to find new replica set")
			return
		}
		if d.toReplicaSet == nil || d.toReplicaSet.ObjectMeta.UID != toReplicaSet.ObjectMeta.UID {
			d.addMessage(fmt.Sprintf("Created new replica set: %s", toReplicaSet.ObjectMeta.Name), "debug")
		}
		d.toReplicaSet = toReplicaSet
		d.db.Child("toUid").Set(toReplicaSet.ObjectMeta.UID)
		d.deployment.ToUID = string(toReplicaSet.ObjectMeta.UID)

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

		rolloutErr := d.monitorRollout()
		if rolloutErr != nil {
			deployment, _, err := kube.Deployments.Get(d.deployment.Env, d.deployment.App)
			if err != nil {
				log.Error(err)
				return
			}
			if deployment != nil {
				deployment.Spec.Paused = true
				_, _, err = kube.Deployments.Replace(d.deployment.Env, d.deployment.App, deployment)
				if err != nil {
					log.Error(err)
					return
				}
			}
			d.db.Child("state").Set(deploymentStatePaused)
			switch e := rolloutErr.(type) {
			case deployerPaused:
				d.addMessage("Paused deployment", "warn")
			case deployerAutoPaused:
				// pass
			case *deployerTimeout:
				d.addMessage(e.message, "error")
			default:
				log.Error(rolloutErr)
			}
		}
	}()
	return nil
}

func (d *deployerSpec) pause() error {
	log.Infof("Pausing deployment %s for app %s in env %s", d.deployment.ID, d.deployment.App, d.deployment.Env)
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
	log.Infof("Rolling back deployment %s for app %s in env %s", d.deployment.ID, d.deployment.App, d.deployment.Env)
	if err := d.acquireLock(); err != nil {
		return err
	}
	defer d.releaseLock()

	var state string
	var newState string
	d.db.Child("state").Value(&state)
	var message string
	switch state {
	case deploymentStatePaused, deploymentStateCompleted:
		message = "Rolling back deployment"
		newState = deploymentStateRollingback
	default:
		return errors.BadRequestError(fmt.Sprintf("Cannot roll back deployment from state %s", state))
	}
	d.db.Child("state").Set(newState)
	d.deployment.State = newState
	d.addMessage(message, "warn")

	deployment, _, err := kube.Deployments.Get(d.deployment.Env, d.deployment.App)
	if err != nil {
		return err
	}
	if deployment.Spec.Paused {
		deployment.Spec.Paused = false
		_, _, err := kube.Deployments.Replace(d.deployment.Env, d.deployment.App, deployment)
		if err != nil {
			return err
		}
	}

	if d.fromReplicaSet != nil && d.toReplicaSet != nil {
		fromRevision, ok := d.fromReplicaSet.ObjectMeta.Annotations["deployment.kubernetes.io/revision"]
		if !ok {
			return errors.New("Unable to determine previous deployment revision")
		}
		rollbackTo, err := strconv.ParseInt(fromRevision, 10, 0)
		if err != nil {
			return err
		}
		rollback := v1beta1.DeploymentRollback{
			Name: d.deployment.App,
			RollbackTo: v1beta1.RollbackConfig{
				Revision: rollbackTo,
			},
		}
		if _, _, err := kube.Deployments.Rollback(d.deployment.Env, d.deployment.App, &rollback); err != nil {
			return err
		}
	} else {
		log.Warn("Cannot roll back deployment with no from or to replica set")
	}
	d.db.Child("state").Set(deploymentStateRolledback)
	d.addMessage("Rolled back deployment", "warn")
	return nil
}

// utils
func (d *deployerSpec) monitorRollout() error {
	for {
		readyCount, runningCount, err := d.refreshReplicaSetPodCount(d.toReplicaSet)
		if err != nil {
			return err
		}
		fromRunningCount := 0
		if d.fromReplicaSet != nil {
			_, fromRunningCount, _ = d.refreshReplicaSetPodCount(d.fromReplicaSet)
		}
		log.Debugf("%s Replica Count: Ready: %v, Running: %v", d.deployment.App, readyCount, runningCount)
		if readyCount == d.deployment.DesiredReplicas && fromRunningCount == 0 {
			d.addMessage(fmt.Sprintf("Successfully completed rollout in %s", d.deployment.Clock.humanize()), "info")
			d.db.Child("state").Set(deploymentStateCompleted)
			return nil
		}
		if err := d.ping(); err != nil {
			return err
		}
		time.Sleep(250 * time.Millisecond)
	}
}

func (d *deployerSpec) acquireLock() error {
	locked, err := redis.GetClient().SetNX(
		fmt.Sprintf("deploymentlock:%s:%s", d.deployment.Env, d.deployment.App),
		true,
		1*time.Hour,
	).Result()
	if err != nil {
		return err
	}
	if !locked {
		return errors.ConflictError("Failed to acquire deployment lock")
	}
	log.Debugf("Acquired lock for %s in %s", d.deployment.App, d.deployment.Env)
	WaitGroup.Add(1)
	return nil
}

func (d *deployerSpec) releaseLock() {
	WaitGroup.Done()
	err := redis.GetClient().Del(
		fmt.Sprintf("deploymentlock:%s:%s", d.deployment.Env, d.deployment.App),
	).Err()
	if err != nil {
		log.Error(err)
	} else {
		log.Debugf("Released lock for %s in %s", d.deployment.App, d.deployment.Env)
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

func (d *deployerSpec) refreshReplicaSetPodCount(replicaSet *v1beta1.ReplicaSet) (int, int, error) {
	kubePods, status, err := kube.Pods.ListForReplicaSet(d.deployment.Env, replicaSet)
	if err != nil {
		return 0, 0, err
	}
	if status != nil {
		return 0, 0, fmt.Errorf("Could not fetch the list of pods for replica set %s", replicaSet.ObjectMeta.Name)
	}
	pods, readyCount, runningCount := getPodsFromPodList(kubePods)

	if string(replicaSet.ObjectMeta.UID) == d.deployment.FromUID {
		d.db.Child("fromPods").Set(pods)
	} else if string(replicaSet.ObjectMeta.UID) == d.deployment.ToUID {
		d.db.Child("toPods").Set(pods)
	}

	return readyCount, runningCount, nil
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
