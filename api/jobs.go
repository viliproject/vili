package api

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/airware/vili/docker"
	"github.com/airware/vili/environments"
	"github.com/airware/vili/log"
	"github.com/airware/vili/templates"
	"gopkg.in/labstack/echo.v1"
)

// JobResponse is the response for the job endpoint
type JobResponse struct {
	Repository  []*docker.Image   `json:"repository,omitempty"`
	PodTemplate string            `json:"podTemplate,omitempty"`
	Variables   map[string]string `json:"variables,omitempty"`
}

func jobHandler(c *echo.Context) error {
	env := c.Param("env")
	job := c.Param("job")

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

	resp := JobResponse{}
	failed := false

	// repository
	var waitGroup sync.WaitGroup
	if len(queryFields) == 0 || queryFields["repository"] {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			branches := []string{"master", "develop"}
			if environment.Branch != "" {
				branches = append(branches, environment.Branch)
			}
			images, err := docker.GetRepository(job, branches)
			if err != nil {
				log.Error(err)
				failed = true
			}
			resp.Repository = images
		}()
	}

	// podTemplate
	if len(queryFields) == 0 || queryFields["podTemplate"] {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			body, err := templates.Pod(environment.Name, environment.Branch, job)
			if err != nil {
				log.Error(err)
				failed = true
			}
			resp.PodTemplate = string(body)
		}()
	}

	// variables
	if len(queryFields) == 0 || queryFields["variables"] {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			variables, err := templates.Variables(environment.Name, environment.Branch)
			if err != nil {
				log.Error(err)
				failed = true
			}
			resp.Variables = variables
		}()
	}

	waitGroup.Wait()
	if failed {
		return fmt.Errorf("failed one of the service calls")
	}

	return c.JSON(http.StatusOK, resp)
}
