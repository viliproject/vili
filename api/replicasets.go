package api

import (
	"net/http"

	"github.com/viliproject/vili/kube"
	"github.com/labstack/echo"
)

func replicaSetsGetHandler(c echo.Context) error {
	env := c.Param("env")

	endpoint := kube.GetClient(env).ReplicaSets()
	query := getListOptionsFromRequest(c)

	if c.Request().URL.Query().Get("watch") != "" {
		return apiWatchWebsocket(c, query, endpoint.Watch)
	}

	// otherwise, return the replicaSets list
	resp, err := endpoint.List(query)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}
