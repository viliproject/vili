package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/airware/vili/errors"
	"github.com/airware/vili/kube"
	"github.com/airware/vili/server"
	"github.com/labstack/echo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

func nodesGetHandler(c echo.Context) error {
	env := c.Param("env")

	endpoint := kube.GetClient(env).Nodes()
	query := getListOptionsFromRequest(c)

	if c.Request().URL.Query().Get("watch") != "" {
		return apiWatchWebsocket(c, query, endpoint.Watch)
	}

	// otherwise, return the nodes list
	resp, err := endpoint.List(query)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

func nodeStateEditHandler(c echo.Context) error {
	env := c.Param("env")
	node := c.Param("node")
	state := c.Param("state")

	endpoint := kube.GetClient(env).Nodes()

	var unschedulable bool
	if state == "enable" {
		unschedulable = false
	} else if state == "disable" {
		unschedulable = true
	} else {
		return server.ErrorResponse(c, errors.BadRequest(fmt.Sprintf("Invalid state for node: %s", state)))
	}

	data := &corev1.Node{
		Spec: corev1.NodeSpec{
			Unschedulable: unschedulable,
		},
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	resp, err := endpoint.Patch(node, types.MergePatchType, dataBytes)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}
