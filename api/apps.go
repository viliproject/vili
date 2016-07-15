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
	"github.com/airware/vili/kube/unversioned"
	"github.com/airware/vili/kube/v1"
	"github.com/airware/vili/log"
	"github.com/airware/vili/server"
	"github.com/airware/vili/templates"
	"github.com/labstack/echo"
)

// AppsResponse is the response for the app endpoint
type AppsResponse struct {
	Controllers *v1.ReplicationControllerList `json:"controllers,omitempty"`
	Services    *v1.ServiceList               `json:"services,omitempty"`
}

func appsHandler(c *echo.Context) error {
	env := c.Param("env")
	log.Info(env)

	resp := AppsResponse{}
	failed := false

	// repository
	var waitGroup sync.WaitGroup

	// controllers
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		controllers, _, err := kube.Controllers.List(env, nil)
		if err != nil {
			log.Error(err)
			failed = true
		}
		resp.Controllers = controllers
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
	Repository         []*docker.Image           `json:"repository,omitempty"`
	ControllerTemplate string                    `json:"controllerTemplate,omitempty"`
	Variables          map[string]string         `json:"variables,omitempty"`
	Controller         *v1.ReplicationController `json:"controller,omitempty"`
	Service            *v1.Service               `json:"service,omitempty"`
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

	// controllerTemplate
	if len(queryFields) == 0 || queryFields["controllerTemplate"] {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			body, err := templates.Controller(env, app)
			if err != nil {
				log.Error(err)
				failed = true
			}
			resp.ControllerTemplate = string(body)
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

	// controller
	if len(queryFields) == 0 || queryFields["controller"] {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			controller, _, err := kube.Controllers.Get(env, app)
			if err != nil {
				log.Error(err)
				failed = true
			}
			resp.Controller = controller
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

	var controllerTemplate templates.Template
	var currentService *v1.Service

	// controllerTemplate
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		body, err := templates.Controller(env, app)
		if err != nil {
			log.Error(err)
			failed = true
		}
		controllerTemplate = body
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
	controller := &v1.ReplicationController{}
	controllerTemplate.Parse(controller)

	controllerPort, err := getPortFromController(controller)
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
					Port:     int32(controllerPort),
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

// ControllerScaleRequest is a request to
type ControllerScaleRequest struct {
	Replicas *int `json:"replicas"`
}

func appScaleHandler(c *echo.Context) error {
	env := c.Param("env")
	app := c.Param("app")

	scaleRequest := &ControllerScaleRequest{}
	decoder := json.NewDecoder(c.Request().Body)
	err := decoder.Decode(scaleRequest)
	if err != nil {
		return err
	}
	if scaleRequest.Replicas == nil {
		return server.ErrorResponse(c, errors.BadRequestError("Replicas missing from scale request"))
	}

	_, status, err := kube.Controllers.Get(env, app)
	if err != nil {
		return err
	}
	if status != nil {
		return fmt.Errorf("Controller %s not found", app)
	}
	replicas := int32(*scaleRequest.Replicas)
	resp, _, err := kube.Controllers.Patch(env, app, &v1.ReplicationController{
		Spec: v1.ReplicationControllerSpec{
			Replicas: &replicas,
		},
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}
