package api

import (
	"net/http"
	"net/url"

	"golang.org/x/net/websocket"

	"github.com/airware/vili/docker"
	"github.com/airware/vili/environments"
	"github.com/airware/vili/kube"
	"github.com/airware/vili/templates"
	echo "gopkg.in/labstack/echo.v1"
)

var (
	jobsQueryParams = []string{"labelSelector", "fieldSelector", "resourceVersion"}
)

func jobsGetHandler(c *echo.Context) error {
	env := c.Param("env")
	query := filterQueryFields(c, jobsQueryParams)

	if c.Request().URL.Query().Get("watch") != "" {
		// watch jobs and return over websocket
		var err error
		websocket.Handler(func(ws *websocket.Conn) {
			err = jobsWatchHandler(ws, env, query)
			ws.Close()
		}).ServeHTTP(c.Response(), c.Request())
		return err
	}

	// otherwise, return the jobs list
	resp, _, err := kube.Jobs.List(env, query)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

func jobsWatchHandler(ws *websocket.Conn, env string, query *url.Values) error {
	return apiWatchHandler(ws, env, query, kube.Jobs.Watch)
}

func jobDeleteHandler(c *echo.Context) error {
	env := c.Param("env")
	job := c.Param("job")

	resp, err := kube.Jobs.Delete(env, job)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

type jobRepositoryResponse struct {
	Images []*docker.Image `json:"images,omitempty"`
}

func jobRepositoryGetHandler(c *echo.Context) error {
	env := c.Param("env")
	job := c.Param("job")

	environment, err := environments.Get(env)
	if err != nil {
		return err
	}

	resp := new(jobRepositoryResponse)
	images, err := docker.GetRepository(job, environment.RepositoryBranches())
	if err != nil {
		return err
	}
	resp.Images = images

	return c.JSON(http.StatusOK, resp)
}

type jobSpecResponse struct {
	Spec string `json:"spec,omitempty"`
}

func jobSpecGetHandler(c *echo.Context) error {
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
