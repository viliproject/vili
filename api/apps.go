package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

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
	"github.com/airware/vili/util"
	"gopkg.in/labstack/echo.v1"
)

// AppsResponse is the response for the app endpoint
type AppsResponse struct {
	ReplicaSets map[string]v1beta1.ReplicaSet `json:"replicaSets,omitempty"`
	Services    *v1.ServiceList               `json:"services,omitempty"`
}

func appsHandler(c *echo.Context) error {
	env := c.Param("env")
	log.Info(env)

	resp := AppsResponse{
		ReplicaSets: make(map[string]v1beta1.ReplicaSet),
	}
	failed := false

	// repository
	var waitGroup sync.WaitGroup

	// deployments
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		deployments, _, err := kube.Deployments.List(env, nil)
		if err != nil {
			log.Error(err)
			failed = true
			return
		}
		waitGroup.Add(len(deployments.Items))
		var rsMutex sync.Mutex
		for _, deployment := range deployments.Items {
			go func(env string, deployment v1beta1.Deployment) {
				defer waitGroup.Done()
				rs, err := getReplicaSetForDeployment(env, &deployment)
				if err != nil {
					log.Error(err)
					failed = true
					return
				}
				rsMutex.Lock()
				resp.ReplicaSets[deployment.Name] = *rs
				rsMutex.Unlock()
			}(env, deployment)
		}
	}()

	// service
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		services, err := kube.Services.List(env)
		if err != nil {
			log.Error(err)
			failed = true
		}
		resp.Services = services
	}()

	waitGroup.Wait()
	if failed {
		return fmt.Errorf("failed one of the service calls")
	}

	return c.JSON(http.StatusOK, resp)
}

// AppResponse is the response for the app endpoint
type AppResponse struct {
	Repository     []*docker.Image     `json:"repository,omitempty"`
	DeploymentSpec string              `json:"deploymentSpec,omitempty"`
	ReplicaSet     *v1beta1.ReplicaSet `json:"replicaSet,omitempty"`
	Service        *v1.Service         `json:"service,omitempty"`
}

func appHandler(c *echo.Context) error {
	env := c.Param("env")
	app := c.Param("app")

	requestFields := c.Request().URL.Query().Get("fields")
	queryFields := make(map[string]bool)
	if requestFields != "" {
		for _, field := range strings.Split(requestFields, ",") {
			queryFields[field] = true
		}
	}

	environment, err := environments.Get(env)
	if err != nil {
		return err
	}

	resp := AppResponse{}
	failed := false

	// repository
	var waitGroup sync.WaitGroup
	if len(queryFields) == 0 || queryFields["repository"] {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			branches := []string{"master", "develop"}
			if environment.Branch != "" && !util.NewStringSet(branches).Contains(environment.Branch) {
				branches = append(branches, environment.Branch)
			}
			images, err := docker.GetRepository(app, branches)
			if err != nil {
				log.Error(err)
				failed = true
			}
			resp.Repository = images
		}()
	}

	// deploymentSpec
	if len(queryFields) == 0 || queryFields["deploymentSpec"] {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			body, err := templates.Deployment(environment.Name, environment.Branch, app)
			if err != nil {
				log.Error(err)
				failed = true
			}
			resp.DeploymentSpec = string(body)
		}()
	}

	// deployment
	if len(queryFields) == 0 || queryFields["deployment"] {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			deployment, _, err := kube.Deployments.Get(env, app)
			if err != nil {
				log.Error(err)
				failed = true
				return
			}
			if deployment == nil {
				return
			}
			rs, err := getReplicaSetForDeployment(env, deployment)
			if err != nil {
				log.Error(err)
				failed = true
				return
			}
			resp.ReplicaSet = rs
		}()
	}

	// service
	if len(queryFields) == 0 || queryFields["service"] {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			service, _, err := kube.Services.Get(env, app)
			if err != nil {
				log.Error(err)
				failed = true
			}
			resp.Service = service
		}()
	}

	waitGroup.Wait()
	if failed {
		return fmt.Errorf("failed one of the service calls")
	}

	return c.JSON(http.StatusOK, resp)
}

func appCreateServiceHandler(c *echo.Context) error {
	env := c.Param("env")
	app := c.Param("app")

	failed := false

	// repository
	var waitGroup sync.WaitGroup

	var deploymentTemplate templates.Template
	var currentService *v1.Service

	environment, err := environments.Get(env)
	if err != nil {
		return err
	}

	// deploymentTemplate
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		body, err := templates.Deployment(environment.Name, environment.Branch, app)
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
		service, _, err := kube.Services.Get(env, app)
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
		return server.ErrorResponse(c, errors.ConflictError("Service exists"))
	}
	deployment := &v1beta1.Deployment{}
	deploymentTemplate.Parse(deployment)

	deploymentPort, err := getPortFromDeployment(deployment)
	if err != nil {
		return err
	}

	service := &v1.Service{
		TypeMeta: unversioned.TypeMeta{
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: app,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				v1.ServicePort{
					Protocol: "TCP",
					Port:     deploymentPort,
				},
			},
			Selector: map[string]string{
				"app": app,
			},
		},
	}

	resp, err := kube.Services.Create(env, app, service)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

// DeploymentScaleRequest is a request to scale a deployment
type DeploymentScaleRequest struct {
	Replicas *int `json:"replicas"`
}

func appScaleHandler(c *echo.Context) error {
	env := c.Param("env")
	app := c.Param("app")

	scaleRequest := &DeploymentScaleRequest{}
	decoder := json.NewDecoder(c.Request().Body)
	err := decoder.Decode(scaleRequest)
	if err != nil {
		return err
	}
	if scaleRequest.Replicas == nil {
		return server.ErrorResponse(c, errors.BadRequestError("Replicas missing from scale request"))
	}

	deployment, status, err := kube.Deployments.Get(env, app)
	if err != nil {
		return err
	}
	if status != nil {
		return fmt.Errorf("Deployment %s not found", app)
	}
	if deployment.Spec.Paused {
		deployment.Spec.Paused = false
		_, _, err := kube.Deployments.Replace(env, app, deployment)
		if err != nil {
			return err
		}
	}
	resp, _, err := kube.Deployments.Scale(env, app, &v1beta1.Scale{
		Spec: v1beta1.ScaleSpec{
			Replicas: int32(*scaleRequest.Replicas),
		},
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

func getReplicaSetForDeployment(env string, deployment *v1beta1.Deployment) (*v1beta1.ReplicaSet, error) {
	replicaSetList, _, err := kube.ReplicaSets.ListForDeployment(env, deployment)
	if err != nil {
		return nil, err
	}
	if replicaSetList == nil {
		return nil, fmt.Errorf("No replicaSet found for deployment %v", *deployment)
	}
	deploymentRevision := deployment.ObjectMeta.Annotations["deployment.kubernetes.io/revision"]
	for _, replicaSet := range replicaSetList.Items {
		rsRevision := replicaSet.ObjectMeta.Annotations["deployment.kubernetes.io/revision"]
		if deploymentRevision == rsRevision {
			return &replicaSet, nil
		}
	}
	return nil, fmt.Errorf("No replicaSet found for deployment %v", *deployment)
}
