package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/airware/vili/config"
	"github.com/airware/vili/docker"
	"github.com/airware/vili/errors"
	"github.com/airware/vili/kube"
	"github.com/airware/vili/kube/extensions/v1beta1"
	"github.com/airware/vili/log"
	"github.com/airware/vili/server"
	"github.com/airware/vili/session"
	"github.com/airware/vili/templates"
	echo "gopkg.in/labstack/echo.v1"
)

const (
	rolloutTimeout = 5 * time.Minute
)

func rolloutCreateHandler(c *echo.Context) error {
	env := c.Param("env")
	deploymentName := c.Param("deployment")

	rollout := new(Rollout)
	if err := json.NewDecoder(c.Request().Body).Decode(rollout); err != nil {
		return err
	}
	if rollout.Branch == "" {
		return server.ErrorResponse(c, errors.BadRequest("Request missing branch"))
	}
	if rollout.Tag == "" {
		return server.ErrorResponse(c, errors.BadRequest("Request missing tag"))
	}
	rollout.Env = env
	rollout.DeploymentName = deploymentName
	rollout.Username = c.Get("user").(*session.User).Username

	err := rollout.Run(c.Request().URL.Query().Get("async") != "")
	if err != nil {
		switch e := err.(type) {
		case RolloutInitError:
			return server.ErrorResponse(c, errors.BadRequest(e.Error()))
		default:
			return e
		}
	}
	return c.JSON(http.StatusOK, rollout)
}

// Rollout represents a single deployment of an image for any app
// TODO: support MaxUnavailable and MaxSurge for rolling updates
type Rollout struct {
	Env            string `json:"env"`
	DeploymentName string `json:"deploymentName"`
	Branch         string `json:"branch"`
	Tag            string `json:"tag"`
	Username       string `json:"username"`
	State          string `json:"state"`

	FromDeployment *v1beta1.Deployment `json:"fromDeployment"`
	FromRevision   string              `json:"fromRevision"`
	ToDeployment   *v1beta1.Deployment `json:"toDeployment"`
	ToRevision     string              `json:"toRevision"`
}

// Run initializes a deployment, checks to make sure it is valid, and runs it
func (r *Rollout) Run(async bool) error {
	digest, err := docker.GetTag(r.DeploymentName, r.Branch, r.Tag)
	if err != nil {
		return err
	}
	if digest == "" {
		return RolloutInitError{
			message: fmt.Sprintf("Tag %s not found for deployment %s", r.Tag, r.DeploymentName),
		}
	}

	r.FromDeployment, _, err = kube.Deployments.Get(r.Env, r.DeploymentName)
	if err != nil {
		return err
	}
	if r.FromDeployment != nil {
		if revision, ok := r.FromDeployment.ObjectMeta.Annotations["deployment.kubernetes.io/revision"]; ok {
			r.FromRevision = revision
		}
	}

	err = r.createNewDeployment()
	if err != nil {
		return err
	}

	if async {
		go r.watchRollout()
		return nil
	}

	return r.watchRollout()
}

func (r *Rollout) createNewDeployment() (err error) {
	// get the spec
	deploymentTemplate, err := templates.Deployment(r.Env, r.Branch, r.DeploymentName)
	if err != nil {
		return
	}

	deployment := new(v1beta1.Deployment)
	err = deploymentTemplate.Parse(deployment)
	if err != nil {
		return
	}

	labels := map[string]string{
		"app":        r.DeploymentName,
		"branch":     r.Branch,
		"deployedBy": r.Username,
	}
	if r.FromRevision != "" {
		labels["fromRevision"] = r.FromRevision
	}
	deployment.ObjectMeta.Labels = labels
	deployment.Spec.Template.ObjectMeta.Labels = labels

	imageName, err := docker.FullName(r.DeploymentName, r.Branch, r.Tag)
	if err != nil {
		return
	}
	deployment.Spec.Template.Spec.Containers[0].Image = imageName

	if r.FromDeployment != nil {
		*deployment.Spec.Replicas = *r.FromDeployment.Spec.Replicas
	}

	deployment.Spec.Strategy.Type = v1beta1.RollingUpdateDeploymentStrategyType

	// create/update deployment
	r.ToDeployment, _, err = kube.Deployments.Replace(r.Env, r.DeploymentName, deployment)
	if err != nil {
		return
	}
	if r.ToDeployment == nil {
		r.ToDeployment, _, err = kube.Deployments.Create(r.Env, deployment)
		if err != nil {
			return
		}
	}

	// wait for ToDeployment to get revision
	err = r.waitRolloutInit()

	r.logMessage(fmt.Sprintf("Rollout for tag %s and branch %s created by %s", r.Tag, r.Branch, r.Username), log.InfoLevel)
	return
}

func (r *Rollout) waitRolloutInit() (err error) {
	watcher, err := kube.Deployments.Watch(r.Env, &url.Values{
		"fieldSelector": {"metadata.name=" + r.DeploymentName},
	})
	if err != nil {
		return err
	}

	startTime := time.Now()
eventLoop:
	for {
		select {
		case event, ok := <-watcher.EventChan:
			if !ok {
				break eventLoop
			}
			if watcher.Stopped() {
				// empty the channel
				continue
			}
			deploymentEvent := event.(*kube.DeploymentEvent)
			finished := false
			r.ToDeployment = deploymentEvent.Object
			if deploymentEvent.List != nil && len(deploymentEvent.List.Items) > 0 {
				r.ToDeployment = &deploymentEvent.List.Items[0]
			}
			switch deploymentEvent.Type {
			case kube.WatchEventDeleted:
				finished = true
			case kube.WatchEventInit, kube.WatchEventAdded, kube.WatchEventModified:
				if revision, ok := r.ToDeployment.ObjectMeta.Annotations["deployment.kubernetes.io/revision"]; ok {
					r.ToRevision = revision
					finished = true
				}
			}
			if finished {
				watcher.Stop()
				break
			}
		case <-time.After(rolloutTimeout):
			elapsed := time.Now().Sub(startTime)
			r.logMessage(fmt.Sprintf("Deployment timed out after %s", humanizeDuration(elapsed)), log.WarnLevel)
			watcher.Stop()
			err = fmt.Errorf("timeout")
			break
		}
	}
	return
}

func (r *Rollout) watchRollout() (err error) {
	watcher, err := kube.Deployments.Watch(r.Env, &url.Values{
		"fieldSelector": {"metadata.name=" + r.DeploymentName},
	})
	if err != nil {
		return err
	}

	startTime := time.Now()
eventLoop:
	for {
		select {
		case event, ok := <-watcher.EventChan:
			if !ok {
				break eventLoop
			}
			if watcher.Stopped() {
				// empty the channel
				continue eventLoop
			}
			elapsed := time.Now().Sub(startTime)
			deploymentEvent := event.(*kube.DeploymentEvent)
			deployment := deploymentEvent.Object
			if deploymentEvent.List != nil && len(deploymentEvent.List.Items) > 0 {
				deployment = &deploymentEvent.List.Items[0]
			}
			switch deploymentEvent.Type {
			case kube.WatchEventDeleted:
				r.logMessage(fmt.Sprintf("Deleted deployment after %s", humanizeDuration(elapsed)), log.WarnLevel)
				watcher.Stop()
				err = fmt.Errorf("deleted")
				break
			case kube.WatchEventInit, kube.WatchEventAdded, kube.WatchEventModified:
				if deployment.Generation <= deployment.Status.ObservedGeneration {
					replicas := *deployment.Spec.Replicas
					if deployment.Status.UpdatedReplicas >= replicas && deployment.Status.AvailableReplicas >= replicas {
						r.logMessage(fmt.Sprintf("Successfully completed rollout in %s", humanizeDuration(elapsed)), log.InfoLevel)
						watcher.Stop()
						break
					}
				}
			}
		case <-time.After(rolloutTimeout):
			elapsed := time.Now().Sub(startTime)
			r.logMessage(fmt.Sprintf("Deployment timed out after %s", humanizeDuration(elapsed)), log.WarnLevel)
			watcher.Stop()
			err = fmt.Errorf("timeout")
			break
		}
	}

	log.Debugf("stopped watching rollout for %s", r.DeploymentName)
	return
}

func (r *Rollout) logMessage(message string, level log.Level) {
	urlStr := fmt.Sprintf(
		"%s/%s/deployments/%s/rollouts",
		config.GetString(config.URI),
		r.Env,
		r.DeploymentName,
	)
	slackMessage := fmt.Sprintf(
		"*%s* - *%s* - <%s|%s> - %s",
		r.Env,
		r.DeploymentName,
		urlStr,
		r.ToRevision,
		message,
	)
	deploymentMessage := fmt.Sprintf(
		"%s - %s - %s",
		r.Env,
		r.DeploymentName,
		message,
	)
	logMessage(deploymentMessage, slackMessage, level)
}

// RolloutInitError is raised if there is a problem initializing a rollout
type RolloutInitError struct {
	message string
}

func (e RolloutInitError) Error() string {
	return e.message
}
