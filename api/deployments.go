package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/CloudCom/firego"
	"github.com/airware/vili/docker"
	"github.com/airware/vili/errors"
	"github.com/airware/vili/firebase"
	"github.com/airware/vili/kube"
	"github.com/airware/vili/kube/v1"
	"github.com/airware/vili/server"
	"github.com/airware/vili/session"
	"github.com/airware/vili/util"
	"gopkg.in/labstack/echo.v1"
)

// Deployment represents a single deployment of an image for any app
type Deployment struct {
	ID       string    `json:"id"`
	Env      string    `json:"env"`
	App      string    `json:"app"`
	Branch   string    `json:"branch"`
	Tag      string    `json:"tag"`
	Time     time.Time `json:"time"`
	Username string    `json:"username"`
	State    string    `json:"state"`
	Rollout  *Rollout  `json:"rollout,omitempty"`

	Clock           *Clock `json:"clock"`
	DesiredReplicas int    `json:"desiredReplicas"`
	OriginalPods    []Pod  `json:"originalPods"`
	FromPods        []Pod  `json:"fromPods"`
	FromTag         string `json:"fromTag"`
	FromUID         string `json:"fromUid"`

	ToPods []Pod  `json:"toPods"`
	ToUID  string `json:"toUid"`
}

// Clock is a time.Duration struct with custom JSON marshal functions
type Clock time.Duration

// MarshalJSON implements the json.Marshaler interface
func (c *Clock) MarshalJSON() ([]byte, error) {
	return json.Marshal(int64(time.Duration(*c) / time.Millisecond))
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (c *Clock) UnmarshalJSON(b []byte) error {
	var ms int64
	err := json.Unmarshal(b, &ms)
	if err != nil {
		return err
	}
	*c = Clock(time.Duration(ms) * time.Millisecond)
	return nil
}
func (c *Clock) humanize() string {
	if c == nil {
		return "0"
	}
	return ((time.Duration(*c) / time.Second) * time.Second).String()
}

// Pod is a summary of the state of a kubernetes pod
type Pod struct {
	Name    string    `json:"name"`
	Created time.Time `json:"created"`
	Phase   string    `json:"phase"`
	Ready   bool      `json:"ready"`
	Host    string    `json:"host"`
}

// Rollout is the description of how the deployment will be rolled out
// TODO: support MaxUnavailable and MaxSurge for rolling updates
type Rollout struct {
	Strategy string `json:"strategy"`
}

const (
	rolloutStrategyRollingUpdate = "RollingUpdate"
	rolloutStrategyRecreate      = "Recreate"
)

const (
	deploymentActionResume   = "resume"
	deploymentActionPause    = "pause"
	deploymentActionRollback = "rollback"
)

const (
	deploymentStateNew         = "new"
	deploymentStateRunning     = "running"
	deploymentStatePausing     = "pausing"
	deploymentStatePaused      = "paused"
	deploymentStateRollingback = "rollingback"
	deploymentStateRolledback  = "rolledback"
	deploymentStateCompleted   = "completed"
)

func deploymentCreateHandler(c *echo.Context) error {
	env := c.Param("env")
	app := c.Param("app")

	deployment := &Deployment{}
	if err := json.NewDecoder(c.Request().Body).Decode(deployment); err != nil {
		return err
	}
	if deployment.Branch == "" {
		return server.ErrorResponse(c, errors.BadRequestError("Request missing branch"))
	}
	if deployment.Tag == "" {
		return server.ErrorResponse(c, errors.BadRequestError("Request missing tag"))
	}
	err := deployment.Init(
		env,
		app,
		c.Get("user").(*session.User).Username,
		c.Request().URL.Query().Get("trigger") != "",
	)
	if err != nil {
		switch e := err.(type) {
		case DeploymentInitError:
			return server.ErrorResponse(c, errors.BadRequestError(e.Error()))
		default:
			return e
		}
	}
	c.JSON(http.StatusOK, deployment)
	return nil
}

func deploymentRolloutEditHandler(c *echo.Context) error {
	env := c.Param("env")
	app := c.Param("app")
	deploymentID := c.Param("deployment")

	rollout := &Rollout{}
	if err := json.NewDecoder(c.Request().Body).Decode(rollout); err != nil {
		return err
	}
	if err := deploymentDB(env, app, deploymentID).Child("rollout").Set(rollout); err != nil {
		return err
	}
	c.JSON(http.StatusOK, rollout)
	return nil
}

func deploymentActionHandler(c *echo.Context) error {
	env := c.Param("env")
	app := c.Param("app")
	deploymentID := c.Param("deployment")
	action := c.Param("action")

	deployment := &Deployment{}
	if err := deploymentDB(env, app, deploymentID).Value(deployment); err != nil {
		return err
	}
	if deployment.ID == "" {
		return server.ErrorResponse(c, errors.NotFoundError("Deployment not found"))
	}
	deployer, err := makeDeployer(env, app, deployment)
	if err != nil {
		return err
	}
	switch action {
	case deploymentActionResume:
		err = deployer.resume()
	case deploymentActionPause:
		err = deployer.pause()
	case deploymentActionRollback:
		err = deployer.rollback()
	default:
		return server.ErrorResponse(c, errors.NotFoundError(fmt.Sprintf("Action %s not found", action)))
	}
	if err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

// utils

// Init initializes a deployment, checks to make sure it is valid, and writes the deployment
// data to firebase
func (d *Deployment) Init(env, app, username string, trigger bool) error {
	d.ID = util.RandLowercaseString(16)
	d.Env = env
	d.App = app
	d.Time = time.Now()
	d.Username = username
	d.State = deploymentStateNew

	digest, err := docker.GetTag(app, d.Branch, d.Tag)
	if err != nil {
		return err
	}
	if digest == "" {
		return DeploymentInitError{
			message: fmt.Sprintf("Tag %s not found for app %s", d.Tag, app),
		}
	}

	deployment, _, err := kube.Deployments.Get(env, app)
	if err != nil {
		return err
	}
	if deployment != nil {
		kubePods, _, err := kube.Pods.ListForDeployment(env, deployment)
		if err != nil {
			return err
		}
		imageTag, err := getImageTagFromDeployment(deployment)
		if err != nil {
			return err
		}
		pods, _, _ := getPodsFromPodList(kubePods)
		if d.DesiredReplicas == 0 {
			d.DesiredReplicas = int(*deployment.Spec.Replicas)
		}
		d.OriginalPods = pods
		d.FromPods = pods
		d.FromTag = imageTag

		replicaSetList, _, _ := kube.ReplicaSets.List(env, &url.Values{
			"labelSelector": []string{"app=" + app},
		})
		revision := deployment.ObjectMeta.Annotations["deployment.kubernetes.io/revision"]
		if replicaSetList != nil {
			for _, replicaSet := range replicaSetList.Items {
				rev := replicaSet.ObjectMeta.Annotations["deployment.kubernetes.io/revision"]
				if rev == revision {
					d.FromUID = string(replicaSet.ObjectMeta.UID)
					break
				}
			}
		}
	}

	if err = deploymentDB(env, app, d.ID).Set(d); err != nil {
		return err
	}

	deployer, err := makeDeployer(env, app, d)
	if err != nil {
		return err
	}

	deployer.addMessage(fmt.Sprintf("Deployment for tag %s and branch %s created by %s", d.Tag, d.Branch, d.Username), "debug")

	if trigger {
		if err := deployer.resume(); err != nil {
			return err
		}
	}

	return nil
}

func deploymentDB(env, app, deploymentID string) *firego.Firebase {
	return firebase.Database().Child(env).Child("apps").Child(app).Child("deployments").Child(deploymentID)
}

func getPodsFromPodList(kubePodList *v1.PodList) (pods []Pod, readyCount, runningCount int) {
	for _, kubePod := range kubePodList.Items {
		pod := Pod{
			Name:    kubePod.ObjectMeta.Name,
			Created: kubePod.ObjectMeta.CreationTimestamp.Time,
			Phase:   string(kubePod.Status.Phase),
		}
		if kubePod.Status.Phase == v1.PodRunning {
			runningCount++
			pod.Ready = true
			for _, containerStatus := range kubePod.Status.ContainerStatuses {
				if !containerStatus.Ready {
					pod.Ready = false
					break
				}
			}
			if pod.Ready {
				readyCount++
			}
		}
		if kubePod.Status.HostIP != "" {
			pod.Host = kubePod.Status.HostIP
		}
		pods = append(pods, pod)
	}
	return
}

// DeploymentInitError is raised if there is a problem initializing a deployment
type DeploymentInitError struct {
	message string
}

func (e DeploymentInitError) Error() string {
	return e.message
}
