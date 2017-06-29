package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"sync"

	"golang.org/x/net/websocket"

	"github.com/airware/vili/docker"
	"github.com/airware/vili/environments"
	"github.com/airware/vili/errors"
	"github.com/airware/vili/kube"
	"github.com/airware/vili/kube/extensions/v1beta1"
	"github.com/airware/vili/kube/unversioned"
	"github.com/airware/vili/kube/v1"
	"github.com/airware/vili/log"
	"github.com/airware/vili/server"
	"github.com/airware/vili/templates"
	echo "gopkg.in/labstack/echo.v1"
)

var (
	deploymentsQueryParams = []string{"labelSelector", "fieldSelector", "resourceVersion"}
)

func deploymentsGetHandler(c *echo.Context) error {
	env := c.Param("env")
	query := filterQueryFields(c, deploymentsQueryParams)

	if c.Request().URL.Query().Get("watch") != "" {
		// watch deployments and return over websocket
		var err error
		websocket.Handler(func(ws *websocket.Conn) {
			err = deploymentsWatchHandler(ws, env, query)
			ws.Close()
		}).ServeHTTP(c.Response(), c.Request())
		return err
	}

	// otherwise, return the deployments list
	resp, _, err := kube.Deployments.List(env, query)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

func deploymentsWatchHandler(ws *websocket.Conn, env string, query *url.Values) error {
	return apiWatchHandler(ws, env, query, kube.Deployments.Watch)
}

type deploymentRepositoryResponse struct {
	Images []*docker.Image `json:"images,omitempty"`
}

func deploymentRepositoryGetHandler(c *echo.Context) error {
	env := c.Param("env")
	deployment := c.Param("deployment")

	environment, err := environments.Get(env)
	if err != nil {
		return err
	}

	resp := new(deploymentRepositoryResponse)
	images, err := docker.GetRepository(deployment, environment.RepositoryBranches)
	if err != nil {
		return err
	}
	resp.Images = images

	return c.JSON(http.StatusOK, resp)
}

type deploymentSpecResponse struct {
	Spec string `json:"spec,omitempty"`
}

func deploymentSpecGetHandler(c *echo.Context) error {
	env := c.Param("env")
	deployment := c.Param("deployment")

	environment, err := environments.Get(env)
	if err != nil {
		return err
	}

	resp := new(deploymentSpecResponse)
	body, err := templates.Deployment(environment.Name, environment.Branch, deployment)
	if err != nil {
		return err
	}
	resp.Spec = string(body)

	return c.JSON(http.StatusOK, resp)
}

func deploymentServiceGetHandler(c *echo.Context) error {
	env := c.Param("env")
	deployment := c.Param("deployment")

	service, _, err := kube.Services.Get(env, deployment)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, service)
}

func deploymentServiceCreateHandler(c *echo.Context) error {
	env := c.Param("env")
	deploymentName := c.Param("deployment")

	failed := false

	var deploymentTemplate templates.Template
	var currentService *v1.Service

	environment, err := environments.Get(env)
	if err != nil {
		return err
	}

	var waitGroup sync.WaitGroup
	// deploymentTemplate
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		body, err := templates.Deployment(environment.Name, environment.Branch, deploymentName)
		if err != nil {
			log.Error(err)
			failed = true
		}
		deploymentTemplate = body
	}()

	// service
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		service, _, err := kube.Services.Get(env, deploymentName)
		if err != nil {
			log.Error(err)
			failed = true
		}
		currentService = service
	}()

	waitGroup.Wait()
	if failed {
		return fmt.Errorf("failed one of the service calls")
	}

	if currentService != nil {
		return server.ErrorResponse(c, errors.Conflict("Service exists"))
	}
	deployment := &v1beta1.Deployment{}
	err = deploymentTemplate.Parse(deployment)
	if err != nil {
		return err
	}

	deploymentPort, err := getPortFromDeployment(deployment)
	if err != nil {
		return err
	}

	service := &v1.Service{
		TypeMeta: unversioned.TypeMeta{
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: deploymentName,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Protocol: "TCP",
					Port:     deploymentPort,
				},
			},
			Selector: map[string]string{
				"app": deploymentName,
			},
		},
	}

	resp, err := kube.Services.Create(env, deploymentName, service)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

type deploymentActionRequest struct {
	Replicas   *int32 `json:"replicas"`
	ToRevision int64  `json:"toRevision"`
}

const (
	deploymentActionResume   = "resume"
	deploymentActionPause    = "pause"
	deploymentActionRollback = "rollback"
	deploymentActionScale    = "scale"
)

func deploymentActionHandler(c *echo.Context) (err error) {
	env := c.Param("env")
	deploymentName := c.Param("deployment")
	action := c.Param("action")

	deployment, status, err := kube.Deployments.Get(env, deploymentName)
	if err != nil {
		return err
	}
	if status != nil {
		return server.ErrorResponse(c, errors.BadRequest(
			fmt.Sprintf("Deployment %s not found", deploymentName)))
	}

	actionRequest := new(deploymentActionRequest)
	// ignore errors, as not all requests have a body
	json.NewDecoder(c.Request().Body).Decode(actionRequest)

	var resp interface{}

	switch action {
	case deploymentActionResume:
		deployment.Spec.Paused = false
		resp, status, err = kube.Deployments.Replace(env, deploymentName, deployment)
	case deploymentActionPause:
		deployment.Spec.Paused = true
		resp, status, err = kube.Deployments.Replace(env, deploymentName, deployment)
	case deploymentActionRollback:
		resp, status, err = kube.Deployments.Rollback(env, deploymentName, &v1beta1.DeploymentRollback{
			Name: deploymentName,
			RollbackTo: v1beta1.RollbackConfig{
				Revision: actionRequest.ToRevision,
			},
		})
	case deploymentActionScale:
		if actionRequest.Replicas == nil {
			return server.ErrorResponse(c, errors.BadRequest("Replicas missing from scale request"))
		}
		resp, status, err = kube.Deployments.Scale(env, deploymentName, &v1beta1.Scale{
			Spec: v1beta1.ScaleSpec{
				Replicas: *actionRequest.Replicas,
			},
		})

	default:
		return server.ErrorResponse(c, errors.NotFound(fmt.Sprintf("Action %s not found", action)))
	}
	if err != nil {
		return err
	}
	if status != nil {
		return c.JSON(http.StatusBadRequest, status)
	}
	return c.JSON(http.StatusOK, resp)
}

func getRolloutHistoryForDeployment(env string, deployment *v1beta1.Deployment) ([]*v1beta1.ReplicaSet, error) {
	replicaSetList, _, err := kube.ReplicaSets.ListForDeployment(env, deployment)
	if err != nil {
		return nil, err
	}
	if replicaSetList == nil {
		return nil, fmt.Errorf("No replicaSet found for deployment %v", *deployment)
	}
	history := byRevision{}
	for _, replicaSet := range replicaSetList.Items {
		rs := replicaSet
		history = append(history, &rs)
	}
	sort.Sort(history)
	return history, nil
}

func getReplicaSetForDeployment(deployment *v1beta1.Deployment, history []*v1beta1.ReplicaSet) (*v1beta1.ReplicaSet, error) {
	deploymentRevision := deployment.ObjectMeta.Annotations["deployment.kubernetes.io/revision"]
	for _, replicaSet := range history {
		rsRevision := replicaSet.ObjectMeta.Annotations["deployment.kubernetes.io/revision"]
		if deploymentRevision == rsRevision {
			return replicaSet, nil
		}
	}
	return nil, fmt.Errorf("No replicaSet found for deployment %v", *deployment)
}

type byRevision []*v1beta1.ReplicaSet

func (s byRevision) Len() int      { return len(s) }
func (s byRevision) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s byRevision) Less(i, j int) bool {
	ri, _ := strconv.Atoi(s[i].ObjectMeta.Annotations["deployment.kubernetes.io/revision"])
	rj, _ := strconv.Atoi(s[j].ObjectMeta.Annotations["deployment.kubernetes.io/revision"])
	return ri > rj
}
