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

func environmentSpecHandler(c *echo.Context) error {
	namespace := c.Query("name")
	branch := c.Query("branch")
	fields := environmentTemplateFields{
		Namespace: namespace,
		Branch:    branch,
	}
	templ, err := templates.Environment(branch)
	if err != nil {
		templ = defaultTemplate
	}
	templ, err = templ.Populate(fields)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]string{
		"spec": string(templ),
	})
}

type environmentTemplateFields struct {
	Namespace string
	Branch    string
}

const defaultTemplate templates.Template = `---
apiVersion: v1
kind: Namespace
metadata:
  name: {{.Namespace}}
  annotations:
    vili.environment-branch: {{.Branch}}
spec:
  finalizers:
  - kubernetes
`
