package api

import (
	"fmt"
	"net/http"
	"net/url"

	"golang.org/x/net/websocket"

	"github.com/airware/vili/errors"
	"github.com/airware/vili/kube"
	"github.com/airware/vili/server"
	echo "gopkg.in/labstack/echo.v1"
)

var (
	nodesQueryParams = []string{"labelSelector", "fieldSelector", "resourceVersion"}
)

func nodesGetHandler(c *echo.Context) error {
	env := c.Param("env")
	query := filterQueryFields(c, nodesQueryParams)

	if c.Request().URL.Query().Get("watch") != "" {
		// watch nodes and return over websocket
		var err error
		websocket.Handler(func(ws *websocket.Conn) {
			err = nodesWatchHandler(ws, env, query)
			ws.Close()
		}).ServeHTTP(c.Response(), c.Request())
		return err
	}

	// otherwise, return the nodes list
	resp, _, err := kube.Nodes.List(env, query)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

func nodesWatchHandler(ws *websocket.Conn, env string, query *url.Values) error {
	return apiWatchHandler(ws, env, query, kube.Nodes.Watch)
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
		return server.ErrorResponse(c, errors.BadRequest(fmt.Sprintf("Invalid state for node: %s", state)))
	}

	resp, err := kube.Nodes.PatchUnschedulable(env, node, unschedulable)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}
