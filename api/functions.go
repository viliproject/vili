package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/airware/vili/environments"
	"github.com/airware/vili/errors"
	"github.com/airware/vili/functions"
	"github.com/airware/vili/repository"
	"github.com/airware/vili/server"
	"github.com/airware/vili/session"
	"github.com/airware/vili/templates"
	"github.com/labstack/echo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
)

func functionsGetHandler(c echo.Context) error {
	env := c.Param("env")

	l, err := functions.List(c.Request().Context(), env)
	if err != nil {
		return err
	}

	sl := []*functionSerializer{}
	for _, f := range l {
		sl = append(sl, makeFunctionSerializer(f))
	}
	if c.Request().URL.Query().Get("watch") != "" {
		return apiWatchWebsocket(c, metav1.ListOptions{}, makeFunctionWatcher(sl))
	}
	return c.JSON(http.StatusOK, sl)
}

func makeFunctionWatcher(functions []*functionSerializer) func(opts metav1.ListOptions) (watch.Interface, error) {
	return func(opts metav1.ListOptions) (watch.Interface, error) {
		eventsChan := make(chan (watch.Event), len(functions))
		for _, function := range functions {
			eventsChan <- watch.Event{
				Type:   watch.Added,
				Object: function,
			}
		}
		return &functionWatcher{
			events:    eventsChan,
			functions: functions,
		}, nil
	}
}

type functionWatcher struct {
	events    chan (watch.Event)
	functions []*functionSerializer
}

func (w *functionWatcher) Stop() {
	close(w.events)
}

func (w *functionWatcher) ResultChan() <-chan watch.Event {
	return w.events
}

type functionRepositoryResponse struct {
	Images []*repository.Image `json:"images,omitempty"`
}

func functionRepositoryGetHandler(c echo.Context) error {
	env := c.Param("env")
	function := c.Param("function")

	environment, err := environments.Get(env)
	if err != nil {
		return err
	}

	codeRepo, err := getFunctionCodeRepo(function, env, environment.RepositoryBranches[0])
	if err != nil {
		return err
	}

	resp := new(functionRepositoryResponse)
	images, err := repository.GetBundleRepository(c.Request().Context(), codeRepo, environment.RepositoryBranches)
	if err != nil {
		return err
	}
	resp.Images = images

	return c.JSON(http.StatusOK, resp)
}

type functionSpecResponse struct {
	Spec string `json:"spec,omitempty"`
}

func functionSpecGetHandler(c echo.Context) error {
	env := c.Param("env")
	function := c.Param("function")

	environment, err := environments.Get(env)
	if err != nil {
		return err
	}

	resp := new(functionSpecResponse)
	body, err := templates.Function(environment.Name, environment.Branch, function)
	if err != nil {
		return err
	}
	resp.Spec = string(body)

	return c.JSON(http.StatusOK, resp)
}

const (
	functionActionDeploy   = "deploy"
	functionActionRollback = "rollback"
)

func functionActionHandler(c echo.Context) (err error) {
	env := c.Param("env")
	functionName := c.Param("function")
	action := c.Param("action")

	_, err = functions.Get(c.Request().Context(), env, functionName)
	if err != nil {
		return err
	}

	switch action {
	case functionActionDeploy:
		return functionDeployHandler(c, env, functionName)
	case functionActionRollback:
		return functionRollbackHandler(c, env, functionName)
	default:
		return server.ErrorResponse(c, errors.NotFound(fmt.Sprintf("Action %s not found", action)))
	}
}

func functionDeployHandler(c echo.Context, env, name string) (err error) {
	spec := new(functions.FunctionDeploySpec)
	if err := json.NewDecoder(c.Request().Body).Decode(spec); err != nil {
		return err
	}
	if spec.Branch == "" {
		return server.ErrorResponse(c, errors.BadRequest("Request missing branch"))
	}
	if spec.Tag == "" {
		return server.ErrorResponse(c, errors.BadRequest("Request missing tag"))
	}
	spec.DeployedBy = c.Get("user").(*session.User).Username

	err = functions.Deploy(c.Request().Context(), env, name, spec)
	if err != nil {
		return
	}
	return c.NoContent(http.StatusNoContent)
}

type functionRollbackRequest struct {
	ToVersion string `json:"toVersion"`
}

func functionRollbackHandler(c echo.Context, env, name string) (err error) {
	rollbackRequest := new(functionRollbackRequest)
	err = json.NewDecoder(c.Request().Body).Decode(rollbackRequest)
	if err != nil {
		return
	}

	err = functions.Rollback(c.Request().Context(), env, name, rollbackRequest.ToVersion)
	if err != nil {
		return
	}
	return c.NoContent(http.StatusNoContent)
}

// functionSerializer is the serializer for functions
type functionSerializer struct {
	Name          string                       `json:"name"`
	ActiveVersion *functionVersionSerializer   `json:"activeVersion"`
	Versions      []*functionVersionSerializer `json:"versions"`
}

// GetObjectKind implements the kubernetes runtime.Object interface
// This allows functions to be returned as part of watch events
func (*functionSerializer) GetObjectKind() schema.ObjectKind {
	return nil
}

// DeepCopyObject implements the kubernetes runtime.Object interface
// This allows functions to be returned as part of watch events
func (*functionSerializer) DeepCopyObject() runtime.Object {
	return nil
}

// functionVersionSerializer is the serializer for function versions
type functionVersionSerializer struct {
	Tag          string    `json:"tag"`
	Branch       string    `json:"branch"`
	Version      string    `json:"version"`
	LastModified time.Time `json:"lastModified"`
	DeployedBy   string    `json:"deployedBy"`
}

func makeFunctionSerializer(function functions.Function) *functionSerializer {
	s := &functionSerializer{
		Name: function.GetName(),
	}
	av := function.GetActiveVersion()
	for _, v := range function.GetVersions() {
		vs := &functionVersionSerializer{
			Tag:          v.GetTag(),
			Branch:       v.GetBranch(),
			Version:      v.GetVersion(),
			LastModified: v.GetLastModified(),
			DeployedBy:   v.GetDeployedBy(),
		}
		s.Versions = append(s.Versions, vs)
		if av != nil && av.GetVersion() == v.GetVersion() {
			s.ActiveVersion = vs
		}
	}
	return s
}
