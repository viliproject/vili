package api

import (
	"encoding/json"
	"net/http"

	"github.com/airware/vili/environments"
	"github.com/labstack/echo"
)

// EnvironmentCreateRequest is a request to create a new environment
type EnvironmentCreateRequest struct {
	Name   string `json:"name"`
	Branch string `json:"branch"`
}

func environmentCreateHandler(c *echo.Context) error {
	envCreateRequest := &EnvironmentCreateRequest{}
	decoder := json.NewDecoder(c.Request().Body)
	err := decoder.Decode(envCreateRequest)
	if err != nil {
		return err
	}

	if envCreateRequest.Name == "" || envCreateRequest.Branch == "" {
		return c.JSON(http.StatusBadRequest, "Must provide a non-empty name and branch")
	}

	if err := environments.Create(envCreateRequest.Name, envCreateRequest.Branch); err != nil {
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
