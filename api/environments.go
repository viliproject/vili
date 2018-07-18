package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/airware/vili/environments"
	"github.com/airware/vili/errors"
	"github.com/airware/vili/session"
	"github.com/airware/vili/templates"
	"github.com/airware/vili/types"
	"github.com/airware/vili/util"
	"github.com/labstack/echo"
)

// EnvironmentsGetResponse is a response to the get environments request
type EnvironmentsGetResponse struct {
	Environments []*environments.Environment `json:"environments"`
}

func environmentsGetHandler(c echo.Context) error {
	envs := environments.Environments()
	if c.QueryParam("branch") != "" {
		allEnvs := envs
		envs = []*environments.Environment{}
		for _, env := range allEnvs {
			if util.NewStringSet(env.AutodeployBranches).Contains(c.QueryParam("branch")) {
				envs = append(envs, env)
			}
		}
	}
	return c.JSON(http.StatusOK, &EnvironmentsGetResponse{
		Environments: envs,
	})
}

// EnvironmentCreateRequest is a request to create a new environment
type EnvironmentCreateRequest struct {
	Name   string `json:"name"`
	Branch string `json:"branch"`
	Spec   string `json:"spec"`
}

// EnvironmentCreateResponse is a response to the create new environment request
type EnvironmentCreateResponse struct {
	Environment *environments.Environment `json:"environment"`
	Resources   map[string][]string       `json:"resources"`
	Release     *types.Release            `json:"release"`
}

func environmentCreateHandler(c echo.Context) error {
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
	environment, err := environments.Get(envCreateRequest.Name)
	if err != nil {
		return err
	}
	release := new(types.Release)
	// get spec for this environment
	spec, err := templates.Release(environment.Name, environment.Branch)
	if err != nil {
		return err
	}
	if err = spec.Parse(release); err != nil {
		return err
	}
	release.Name = "init"
	release.TargetEnv = environment.Name
	release.CreatedAt = time.Now()
	release.CreatedBy = c.Get("user").(*session.User).Username
	if populateReleaseLatestVersions(c.Request().Context(), environment, release) {
		return errors.InternalServerError()
	}
	// save release to the database
	err = setReleaseValue(release)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, &EnvironmentCreateResponse{
		Environment: environment,
		Resources:   resources,
		Release:     release,
	})
}

func environmentDeleteHandler(c echo.Context) error {
	env := c.Param("env")

	if err := environments.Delete(env); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func environmentSpecHandler(c echo.Context) error {
	namespace := c.QueryParam("name")
	branch := c.QueryParam("branch")
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
