package api

import (
	"net/http"
	"net/url"

	"github.com/airware/vili/kube"
	"golang.org/x/net/websocket"
	echo "gopkg.in/labstack/echo.v1"
)

var (
	replicaSetsQueryParams = []string{"labelSelector", "fieldSelector", "resourceVersion"}
)

func replicaSetsGetHandler(c *echo.Context) error {
	env := c.Param("env")
	query := filterQueryFields(c, replicaSetsQueryParams)

	if c.Request().URL.Query().Get("watch") != "" {
		// watch replicaSets and return over websocket
		var err error
		websocket.Handler(func(ws *websocket.Conn) {
			err = replicaSetsWatchHandler(ws, env, query)
			ws.Close()
		}).ServeHTTP(c.Response(), c.Request())
		return err
	}

	// otherwise, return the replicaSets list
	resp, _, err := kube.ReplicaSets.List(env, query)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

func replicaSetsWatchHandler(ws *websocket.Conn, env string, query *url.Values) error {
	return apiWatchHandler(ws, env, query, kube.ReplicaSets.Watch)
}
