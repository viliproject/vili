package api

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/airware/vili/environments"
	"github.com/airware/vili/errors"
	"github.com/airware/vili/kube"
	"github.com/airware/vili/kube/v1"
	"github.com/airware/vili/templates"
	"golang.org/x/net/websocket"
	echo "gopkg.in/labstack/echo.v1"
)

var (
	configmapsQueryParams = []string{"labelSelector", "fieldSelector", "resourceVersion"}
)

func configmapsGetHandler(c *echo.Context) error {
	env := c.Param("env")
	query := filterQueryFields(c, configmapsQueryParams)

	if c.Request().URL.Query().Get("watch") != "" {
		// watch configmaps and return over websocket
		var err error
		websocket.Handler(func(ws *websocket.Conn) {
			err = configmapsWatchHandler(ws, env, query)
			ws.Close()
		}).ServeHTTP(c.Response(), c.Request())
		return err
	}

	// otherwise, return the configmaps list
	resp, _, err := kube.ConfigMaps.List(env, query)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

func configmapsWatchHandler(ws *websocket.Conn, env string, query *url.Values) error {
	return apiWatchHandler(ws, env, query, kube.ConfigMaps.Watch)
}

func configmapSpecGetHandler(c *echo.Context) error {
	env := c.Param("env")
	configmapName := c.Param("configmap")

	environment, err := environments.Get(env)
	if err != nil {
		return err
	}

	configmapTemplate, err := templates.ConfigMap(environment.Name, environment.Branch, configmapName)
	if err != nil {
		return err
	}
	configmap := new(v1.ConfigMap)
	err = configmapTemplate.Parse(configmap)
	return c.JSON(http.StatusOK, configmap)
}

func configmapCreateHandler(c *echo.Context) error {
	env := c.Param("env")
	configmapName := c.Param("configmap")

	environment, err := environments.Get(env)
	if err != nil {
		return err
	}

	configmapTemplate, err := templates.ConfigMap(environment.Name, environment.Branch, configmapName)
	if err != nil {
		return err
	}
	configmap := new(v1.ConfigMap)
	err = configmapTemplate.Parse(configmap)

	configmap, resp, err := kube.ConfigMaps.Create(env, configmap)
	if err != nil {
		return err
	}
	if resp != nil {
		return c.JSON(http.StatusBadRequest, resp)
	}
	return c.JSON(http.StatusOK, configmap)
}

func configmapDeleteHandler(c *echo.Context) error {
	env := c.Param("env")
	configmapName := c.Param("configmap")

	resp, err := kube.ConfigMaps.Delete(env, configmapName)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

func configmapSetKeysHandler(c *echo.Context) error {
	env := c.Param("env")
	configmapName := c.Param("configmap")

	data := map[string]string{}
	err := json.NewDecoder(c.Request().Body).Decode(&data)
	if err != nil {
		return errors.BadRequest("Invalid body")
	}

	configmap, resp, err := kube.ConfigMaps.Get(env, configmapName)
	if err != nil {
		return err
	}
	if resp != nil {
		return c.JSON(http.StatusBadRequest, resp)
	}
	for key, val := range data {
		configmap.Data[key] = val
	}
	configmap, _, err = kube.ConfigMaps.Replace(env, configmapName, configmap)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, configmap)
}

func configmapDeleteKeyHandler(c *echo.Context) error {
	env := c.Param("env")
	configmapName := c.Param("configmap")
	key := c.Param("key")

	configmap, _, err := kube.ConfigMaps.Get(env, configmapName)
	if err != nil {
		return err
	}

	if _, ok := configmap.Data[key]; ok {
		delete(configmap.Data, key)
	}
	resp, _, err := kube.ConfigMaps.Replace(env, configmapName, configmap)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}
