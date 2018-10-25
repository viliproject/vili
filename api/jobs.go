package api

import (
	"net/http"

	"github.com/viliproject/vili/environments"
	"github.com/viliproject/vili/kube"
	"github.com/viliproject/vili/repository"
	"github.com/viliproject/vili/templates"
	"github.com/labstack/echo"
)

func jobsGetHandler(c echo.Context) error {
	env := c.Param("env")

	endpoint := kube.GetClient(env).Jobs()
	query := getListOptionsFromRequest(c)

	if c.Request().URL.Query().Get("watch") != "" {
		return apiWatchWebsocket(c, query, endpoint.Watch)
	}

	// otherwise, return the jobs list
	resp, err := endpoint.List(query)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

func jobDeleteHandler(c echo.Context) error {
	env := c.Param("env")
	job := c.Param("job")

	endpoint := kube.GetClient(env).Jobs()

	err := endpoint.Delete(job, nil)
	if err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

type jobRepositoryResponse struct {
	Images []*repository.Image `json:"images,omitempty"`
}

func jobRepositoryGetHandler(c echo.Context) error {
	env := c.Param("env")
	job := c.Param("job")

	environment, err := environments.Get(env)
	if err != nil {
		return err
	}

	resp := new(jobRepositoryResponse)
	images, err := repository.GetDockerRepository(job, environment.RepositoryBranches)
	if err != nil {
		return err
	}
	resp.Images = images

	return c.JSON(http.StatusOK, resp)
}

type jobSpecResponse struct {
	Spec string `json:"spec,omitempty"`
}

func jobSpecGetHandler(c echo.Context) error {
	env := c.Param("env")
	job := c.Param("job")

	environment, err := environments.Get(env)
	if err != nil {
		return err
	}

	resp := new(jobSpecResponse)
	body, err := templates.Job(environment.Name, environment.Branch, job)
	if err != nil {
		return err
	}
	resp.Spec = string(body)

	return c.JSON(http.StatusOK, resp)
}
