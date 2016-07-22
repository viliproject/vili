package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/CloudCom/firego"
	"github.com/airware/vili/docker"
	"github.com/airware/vili/errors"
	"github.com/airware/vili/firebase"
	"github.com/airware/vili/log"
	"github.com/airware/vili/server"
	"github.com/airware/vili/session"
	"github.com/airware/vili/templates"
	"github.com/airware/vili/util"
	"github.com/labstack/echo"
)

// Run represents a single run of an image for any job
type Run struct {
	ID       string    `json:"id"`
	Branch   string    `json:"branch"`
	Tag      string    `json:"tag"`
	Time     time.Time `json:"time"`
	Username string    `json:"username"`
	State    string    `json:"state"`

	Clock *Clock `json:"clock"`

	UID string `json:"uid"`

	PodTemplate templates.Template `json:"template"`
	Variables   map[string]string  `json:"variables"`
}

const (
	runActionStart     = "start"
	runActionTerminate = "terminate"
)

const (
	runStateNew         = "new"
	runStateRunning     = "running"
	runStateTerminating = "terminating"
	runStateTerminated  = "terminated"
	runStateCompleted   = "completed"
	runStateFailed      = "failed"
)

func runCreateHandler(c *echo.Context) error {
	env := c.Param("env")
	job := c.Param("job")

	run := &Run{}
	if err := json.NewDecoder(c.Request().Body).Decode(run); err != nil {
		return err
	}
	if run.Tag == "" {
		return server.ErrorResponse(c, errors.BadRequestError("Request missing tag"))
	}

	err := run.Init(
		env,
		job,
		c.Get("user").(*session.User).Username,
		c.Request().URL.Query().Get("trigger") != "",
	)
	if err != nil {
		switch e := err.(type) {
		case RunInitError:
			return server.ErrorResponse(c, errors.BadRequestError(e.Error()))
		default:
			return e
		}
	}
	c.JSON(http.StatusOK, run)
	return nil
}

func runActionHandler(c *echo.Context) error {
	env := c.Param("env")
	job := c.Param("job")
	runID := c.Param("run")
	action := c.Param("action")

	run := &Run{}
	if err := runDB(env, job, runID).Value(run); err != nil {
		return err
	}
	if run.ID == "" {
		return server.ErrorResponse(c, errors.NotFoundError("Run not found"))
	}
	log.Info(run.ID)
	log.Info(run.Clock)

	runner, err := makeRunner(env, job, run)
	if err != nil {
		return err
	}
	switch action {
	case runActionStart:
		err = runner.start()
	case runActionTerminate:
		err = runner.terminate()
	default:
		return server.ErrorResponse(c, errors.NotFoundError(fmt.Sprintf("Action %s not found", action)))
	}
	if err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func runVariablesEditHandler(c *echo.Context) error {
	env := c.Param("env")
	job := c.Param("job")
	runID := c.Param("run")

	run := &Run{}
	if err := runDB(env, job, runID).Value(run); err != nil {
		return err
	}
	if run.ID == "" {
		return server.ErrorResponse(c, errors.NotFoundError("Run not found"))
	}
	if run.State != "new" {
		return server.ErrorResponse(c, errors.BadRequestError("Can only update variables for new jobs"))
	}

	variables := make(map[string]string)
	if err := json.NewDecoder(c.Request().Body).Decode(&variables); err != nil {
		return err
	}

	if err := runDB(env, job, runID).Child("variables").Set(variables); err != nil {
		return err
	}

	c.JSON(http.StatusOK, variables)
	return nil
}

// utils

// Init initializes a job run, checks to make sure it is valid, and writes the run
// data to firebase
func (r *Run) Init(env, job, username string, trigger bool) error {
	r.ID = util.RandLowercaseString(16)
	r.Time = time.Now()
	r.Username = username
	r.State = runStateNew

	digest, err := docker.GetTag(job, r.Branch, r.Tag)
	if err != nil {
		return err
	}
	if digest == "" {
		return RunInitError{
			message: fmt.Sprintf("Tag %s not found for job %s", r.Tag, job),
		}
	}

	body, err := templates.Pod(env, job)
	if err != nil {
		return err
	}
	r.PodTemplate = body
	r.Variables = r.PodTemplate.ExtractVariables()

	if err = runDB(env, job, r.ID).Set(r); err != nil {
		return err
	}

	runner, err := makeRunner(env, job, r)
	if err != nil {
		return err
	}
	runner.addMessage(fmt.Sprintf("Job run for tag %s created by %s", r.Tag, r.Username), "debug")

	if trigger {
		if err := runner.start(); err != nil {
			return err
		}
	}

	return nil
}

func runDB(env, job, runID string) *firego.Firebase {
	return firebase.Database().Child(env).Child("jobs").Child(job).Child("runs").Child(runID)
}

// RunInitError is raised if there is a problem initializing a run
type RunInitError struct {
	message string
}

func (e RunInitError) Error() string {
	return e.message
}
