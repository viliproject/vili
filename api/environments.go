package api

import (
	"net/http"

	"github.com/airware/vili/environments"
	"github.com/labstack/echo"
)

func environmentCreateHandler(c *echo.Context) error {
	env := c.Param("env")

	if err := environments.Create(env); err != nil {
		return err
	}
	return c.NoContent(http.StatusCreated)
}

func environmentDeleteHandler(c *echo.Context) error {
	env := c.Param("env")

	if err := environments.Delete(env); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}
