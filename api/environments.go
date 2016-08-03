package api

import (
	"encoding/json"
	"net/http"

	"github.com/airware/vili/environments"
	"github.com/airware/vili/templates"
	"gopkg.in/labstack/echo.v1"
)

// EnvironmentCreateRequest is a request to create a new environment
type EnvironmentCreateRequest struct {
	Name   string `json:"name"`
	Branch string `json:"branch"`
	Spec   string `json:"spec"`
}

func environmentCreateHandler(c *echo.Context) error {
	envCreateRequest := &EnvironmentCreateRequest{}
	decoder := json.NewDecoder(c.Request().Body)
	err := decoder.Decode(envCreateRequest)
	if err != nil {
		return err
	}

	if envCreateRequest.Name == "" || envCreateRequest.Branch == "" || envCreateRequest.Spec == "" {
		return c.JSON(http.StatusBadRequest, "Must provide a non-empty name, branch, and spec")
	}

	resources, err := environments.Create(envCreateRequest.Name, envCreateRequest.Branch, envCreateRequest.Spec)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, resources)
}

func environmentDeleteHandler(c *echo.Context) error {
	env := c.Param("env")

	if err := environments.Delete(env); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func environmentTemplateHandler(c *echo.Context) error {
	branch := c.Query("branch")
	if branch == "" {
		branch = "master"
	}

	templ, err := templates.Environment(branch)
	if err != nil {
		// Fall back to the main branch before returning a basic template
		templ, err = templates.Environment("")
		if err != nil {
			return c.JSON(http.StatusOK, map[string]string{
				"template": defaultTemplate,
				"details":  err.Error(),
			})
		}
	}
	return c.JSON(http.StatusOK, map[string]string{
		"template": string(templ),
	})
}

const defaultTemplate string = `---
apiVersion: v1
kind: Namespace
metadata:
  name: {NAMESPACE}
  annotations:
    vili.environment-branch: {BRANCH}
spec:
  finalizers:
  - kubernetes
`
