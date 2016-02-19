package api

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/airware/vili/errors"
	"github.com/airware/vili/kube"
	"github.com/airware/vili/server"
	"github.com/labstack/echo"
)

var nodesQueryParams = []string{"labelSelector", "fieldSelector"}

func nodesHandler(c *echo.Context) error {
	env := c.Param("env")

	query := &url.Values{}
	for _, param := range nodesQueryParams {
		val := c.Request().URL.Query().Get(param)
		if val != "" {
			query.Add(param, val)
		}
	}
	resp, err := kube.Nodes.List(env, query)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

func nodeHandler(c *echo.Context) error {
	env := c.Param("env")
	node := c.Param("node")

	resp, status, err := kube.Nodes.Get(env, node)
	if err != nil {
		return err
	}
	if status != nil {
		return c.JSON(http.StatusOK, status)
	}
	return c.JSON(http.StatusOK, resp)
}

func nodeStateEditHandler(c *echo.Context) error {
	env := c.Param("env")
	node := c.Param("node")
	state := c.Param("state")

	var unschedulable bool
	if state == "enable" {
		unschedulable = false
	} else if state == "disable" {
		unschedulable = true
	} else {
		return server.ErrorResponse(c, errors.BadRequestError(fmt.Sprintf("Invalid state for node: %s", state)))
	}

	resp, err := kube.Nodes.PatchUnschedulable(env, node, unschedulable)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}
