package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/airware/vili/config"
	"github.com/airware/vili/errors"
	"github.com/airware/vili/kube"
	"github.com/airware/vili/log"
	"github.com/airware/vili/repository"
	"github.com/airware/vili/server"
	"github.com/airware/vili/session"
	"github.com/labstack/echo"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	kubeErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

func rolloutCreateHandler(c echo.Context) error {
	req := c.Request()
	env := c.Param("env")
	deploymentName := c.Param("deployment")

	rollout := new(Rollout)
	if err := json.NewDecoder(req.Body).Decode(rollout); err != nil {
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

	err := rollout.Run(req.Context(), req.URL.Query().Get("async") != "")
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

	FromDeployment *extv1beta1.Deployment `json:"fromDeployment"`
	FromRevision   string                 `json:"fromRevision"`
	ToDeployment   *extv1beta1.Deployment `json:"toDeployment"`
	ToRevision     string                 `json:"toRevision"`
}

// Run initializes a deployment, checks to make sure it is valid, and runs it
func (r *Rollout) Run(ctx context.Context, async bool) error {
	err := r.createNewDeployment(ctx)
	if err != nil {
		return err
	}

	if async {
		go r.watchRollout()
		return nil
	}

	return r.watchRollout()
}

func (r *Rollout) createNewDeployment(ctx context.Context) (err error) {
	// get the spec
	deployment, err := getDeploymentWithTag(r.Env, r.Branch, r.DeploymentName, r.Tag)
	if err != nil {
		return err
	}

	imageRepo, err := getImageRepoFromDeployment(deployment)
	if err != nil {
		return err
	}

	// check to see if tag is valid
	digest, err := repository.GetDockerTag(ctx, imageRepo, r.Tag)
	if err != nil {
		return err
	}
	if digest == "" {
		return RolloutInitError{
			message: fmt.Sprintf("Tag %s not found for deployment %s", r.Tag, r.DeploymentName),
		}
	}

	// get previous deployment
	fromDeployment, err := kube.GetClient(r.Env).Deployments().Get(r.DeploymentName, metav1.GetOptions{})
	if err != nil {
		if statusError, ok := err.(*kubeErrors.StatusError); !ok || statusError.Status().Code != http.StatusNotFound {
			// only return error if the error is something other than NotFound
			return err
		}
	} else {
		r.FromDeployment = fromDeployment
		if revision, ok := r.FromDeployment.ObjectMeta.Annotations["deployment.kubernetes.io/revision"]; ok {
			r.FromRevision = revision
		}
	}

	// add labels
	labels := map[string]string{
		"app": r.DeploymentName,
	}
	deployment.ObjectMeta.Name = r.DeploymentName
	deployment.ObjectMeta.Labels = labels
	deployment.Spec.Template.ObjectMeta.Labels = labels

	// add annotations
	if deployment.ObjectMeta.Annotations == nil {
		deployment.ObjectMeta.Annotations = map[string]string{}
	}
	if deployment.Spec.Template.ObjectMeta.Annotations == nil {
		deployment.Spec.Template.ObjectMeta.Annotations = map[string]string{}
	}
	deployment.ObjectMeta.Annotations["vili/branch"] = r.Branch
	deployment.Spec.Template.ObjectMeta.Annotations["vili/branch"] = r.Branch
	deployment.ObjectMeta.Annotations["vili/deployedBy"] = r.Username
	deployment.Spec.Template.ObjectMeta.Annotations["vili/deployedBy"] = r.Username
	if r.FromRevision != "" {
		deployment.ObjectMeta.Annotations["vili/fromRevision"] = r.FromRevision
		deployment.Spec.Template.ObjectMeta.Annotations["vili/fromRevision"] = r.FromRevision
	}

	if r.FromDeployment != nil {
		*deployment.Spec.Replicas = *r.FromDeployment.Spec.Replicas
	}

	deployment.Spec.Strategy.Type = extv1beta1.RollingUpdateDeploymentStrategyType

	// create/update deployment
	endpoint := kube.GetClient(r.Env).Deployments()
	r.ToDeployment, err = endpoint.Update(deployment)
	if err != nil {
		if statusError, ok := err.(*kubeErrors.StatusError); ok && statusError.Status().Code == http.StatusNotFound {
			r.ToDeployment, err = endpoint.Create(deployment)
			if err != nil {
				return
			}
		} else {
			return
		}
	}

	// wait for ToDeployment to get revision
	err = r.waitRolloutInit()

	r.logMessage(fmt.Sprintf("Rollout for tag %s and branch %s created by %s", r.Tag, r.Branch, r.Username), log.InfoLevel)
	return
}

func (r *Rollout) waitRolloutInit() (err error) {
	watcher, err := kube.GetClient(r.Env).Deployments().Watch(metav1.ListOptions{
		FieldSelector: "metadata.name=" + r.DeploymentName,
	})
	if err != nil {
		return err
	}

	startTime := time.Now()
eventLoop:
	for {
		select {
		case event, ok := <-watcher.ResultChan():
			if !ok {
				break eventLoop
			}
			r.ToDeployment = event.Object.(*extv1beta1.Deployment)
			finished := false
			switch event.Type {
			case watch.Deleted:
				finished = true
			case watch.Added, watch.Modified:
				if revision, ok := r.ToDeployment.ObjectMeta.Annotations["deployment.kubernetes.io/revision"]; ok {
					r.ToRevision = revision
					finished = true
				}
			}
			if finished {
				watcher.Stop()
				break eventLoop
			}
		case <-time.After(config.GetDuration(config.RolloutTimeout)):
			elapsed := time.Now().Sub(startTime)
			r.logMessage(fmt.Sprintf("Deployment timed out after %s", humanizeDuration(elapsed)), log.WarnLevel)
			watcher.Stop()
			err = fmt.Errorf("timeout")
			break eventLoop
		}
	}
	return
}

func (r *Rollout) watchRollout() (err error) {
	watcher, err := kube.GetClient(r.Env).Deployments().Watch(metav1.ListOptions{
		FieldSelector: "metadata.name=" + r.DeploymentName,
	})
	if err != nil {
		return err
	}

	startTime := time.Now()
eventLoop:
	for {
		select {
		case event, ok := <-watcher.ResultChan():
			if !ok {
				break eventLoop
			}
			elapsed := time.Now().Sub(startTime)
			deployment := event.Object.(*extv1beta1.Deployment)
			switch event.Type {
			case watch.Deleted:
				r.logMessage(fmt.Sprintf("Deleted deployment after %s", humanizeDuration(elapsed)), log.WarnLevel)
				watcher.Stop()
				err = fmt.Errorf("deleted")
				break eventLoop
			case watch.Added, watch.Modified:
				if deployment.Generation <= deployment.Status.ObservedGeneration {
					replicas := *deployment.Spec.Replicas
					if deployment.Status.UpdatedReplicas >= replicas && deployment.Status.AvailableReplicas >= replicas {
						r.logMessage(fmt.Sprintf("Successfully completed rollout in %s", humanizeDuration(elapsed)), log.InfoLevel)
						watcher.Stop()
						break eventLoop
					}
				}
			}
		case <-time.After(config.GetDuration(config.RolloutTimeout)):
			elapsed := time.Now().Sub(startTime)
			r.logMessage(fmt.Sprintf("Deployment timed out after %s", humanizeDuration(elapsed)), log.WarnLevel)
			watcher.Stop()
			err = fmt.Errorf("timeout")
			break eventLoop
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
