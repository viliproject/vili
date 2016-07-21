package api

import (
	"net/http"

	"github.com/airware/vili/config"
	"github.com/airware/vili/kube"
	"github.com/airware/vili/kube/v1"
	"github.com/airware/vili/util"
	"github.com/labstack/echo"
)

func environmentCreateHandler(c *echo.Context) error {
	env := c.Param("env")

	resp, status, err := kube.Namespaces.Create(&v1.Namespace{
		ObjectMeta: v1.ObjectMeta{
			Name: env,
		},
	})
	if err != nil {
		return err
	}
	if status != nil {
		return c.JSON(http.StatusOK, status)
	}
	return c.JSON(http.StatusOK, resp)
}

func environmentDeleteHandler(c *echo.Context) error {
	env := c.Param("env")

	protectedEnvs := util.NewStringSet(config.GetStringSlice(config.Environments))
	if protectedEnvs.Contains(env) {
		return c.JSON(http.StatusBadRequest, env+" is a protected environment")
	}

	status, err := kube.Namespaces.Delete(env)
	if err != nil {
		return err
	}
	if status != nil {
		return c.JSON(http.StatusOK, status)
	}
	return c.NoContent(http.StatusNoContent)
}
