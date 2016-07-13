package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/airware/vili/docker"
	"github.com/airware/vili/errors"
	"github.com/airware/vili/kube"
	"github.com/airware/vili/kube/extensions/v1beta1"
	"github.com/airware/vili/kube/unversioned"
	"github.com/airware/vili/kube/v1"
	"github.com/airware/vili/log"
	"github.com/airware/vili/server"
	"github.com/airware/vili/templates"
	"github.com/labstack/echo"
)

// AppsResponse is the response for the app endpoint
type AppsResponse struct {
	Deployments *v1beta1.DeploymentList `json:"deployments,omitempty"`
	Services    *v1.ServiceList         `json:"services,omitempty"`
}

func appsHandler(c *echo.Context) error {
	env := c.Param("env")
	log.Info(env)

	resp := AppsResponse{}
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
		}
		resp.Deployments = deployments
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
	Repository         []*docker.Image     `json:"repository,omitempty"`
	DeploymentTemplate string              `json:"deploymentTemplate,omitempty"`
	Variables          map[string]string   `json:"variables,omitempty"`
	Deployment         *v1beta1.Deployment `json:"deployment,omitempty"`
	Service            *v1.Service         `json:"service,omitempty"`
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

	resp := AppResponse{}
	failed := false

	// repository
	var waitGroup sync.WaitGroup
	if len(queryFields) == 0 || queryFields["repository"] {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			images, err := docker.GetRepository(app, true)
			if err != nil {
				log.Error(err)
				failed = true
			}
			resp.Repository = images
		}()
	}

	// deploymentTemplate
	if len(queryFields) == 0 || queryFields["deploymentTemplate"] {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			body, err := templates.Deployment(env, app)
			if err != nil {
				log.Error(err)
				failed = true
			}
			resp.DeploymentTemplate = string(body)
		}()
	}

	// variables
	if len(queryFields) == 0 || queryFields["variables"] {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			variables, err := templates.Variables(env)
			if err != nil {
				log.Error(err)
				failed = true
			}
			resp.Variables = variables
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
			}
			resp.Deployment = deployment
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

	// deploymentTemplate
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		body, err := templates.Deployment(env, app)
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

// DeploymentScaleRequest is a request to
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
