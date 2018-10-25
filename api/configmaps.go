package api

import (
	"encoding/json"
	"net/http"

	"github.com/viliproject/vili/environments"
	"github.com/viliproject/vili/errors"
	"github.com/viliproject/vili/kube"
	"github.com/viliproject/vili/templates"
	"github.com/labstack/echo"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	configmapsQueryParams = []string{"labelSelector", "fieldSelector", "resourceVersion"}
)

func configmapsGetHandler(c echo.Context) error {
	env := c.Param("env")

	endpoint := kube.GetClient(env).ConfigMaps()
	query := getListOptionsFromRequest(c)

	if c.Request().URL.Query().Get("watch") != "" {
		return apiWatchWebsocket(c, query, endpoint.Watch)
	}

	// otherwise, return the configmaps list
	resp, err := endpoint.List(query)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

func configmapSpecGetHandler(c echo.Context) error {
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
	configmap := new(corev1.ConfigMap)
	err = configmapTemplate.Parse(configmap)
	return c.JSON(http.StatusOK, configmap)
}

func configmapCreateHandler(c echo.Context) error {
	env := c.Param("env")
	configmapName := c.Param("configmap")

	endpoint := kube.GetClient(env).ConfigMaps()

	environment, err := environments.Get(env)
	if err != nil {
		return err
	}

	configmapTemplate, err := templates.ConfigMap(environment.Name, environment.Branch, configmapName)
	if err != nil {
		return err
	}
	configmap := new(corev1.ConfigMap)
	err = configmapTemplate.Parse(configmap)

	configmap, err = endpoint.Create(configmap)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, configmap)
}

func configmapDeleteHandler(c echo.Context) error {
	env := c.Param("env")
	configmapName := c.Param("configmap")

	endpoint := kube.GetClient(env).ConfigMaps()

	err := endpoint.Delete(configmapName, nil)
	if err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent) // TODO return status?
}

func configmapSetKeysHandler(c echo.Context) error {
	env := c.Param("env")
	configmapName := c.Param("configmap")

	endpoint := kube.GetClient(env).ConfigMaps()

	data := map[string]string{}
	err := json.NewDecoder(c.Request().Body).Decode(&data)
	if err != nil {
		return errors.BadRequest("Invalid body")
	}

	configmap, err := endpoint.Get(configmapName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	for key, val := range data {
		configmap.Data[key] = val
	}
	configmap, err = endpoint.Update(configmap)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, configmap)
}

func configmapDeleteKeyHandler(c echo.Context) error {
	env := c.Param("env")
	configmapName := c.Param("configmap")
	key := c.Param("key")

	endpoint := kube.GetClient(env).ConfigMaps()

	configmap, err := endpoint.Get(configmapName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if _, ok := configmap.Data[key]; ok {
		delete(configmap.Data, key)
	}
	resp, err := endpoint.Update(configmap)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}
