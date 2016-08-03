package api

import (
	"net/http"
	"net/url"

	"github.com/airware/vili/kube"
	"gopkg.in/labstack/echo.v1"
)

var podsQueryParams = []string{"labelSelector", "fieldSelector"}

func podsHandler(c *echo.Context) error {
	env := c.Param("env")

	query := &url.Values{}
	for _, param := range podsQueryParams {
		val := c.Request().URL.Query().Get(param)
		if val != "" {
			query.Add(param, val)
		}
	}
	resp, _, err := kube.Pods.List(env, query)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

func podHandler(c *echo.Context) error {
	env := c.Param("env")
	pod := c.Param("pod")

	resp, status, err := kube.Pods.Get(env, pod)
	if err != nil {
		return err
	}
	if status != nil {
		return c.JSON(http.StatusOK, status)
	}
	return c.JSON(http.StatusOK, resp)
}

func podDeleteHandler(c *echo.Context) error {
	env := c.Param("env")
	pod := c.Param("pod")

	resp, _, err := kube.Pods.Delete(env, pod)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}
