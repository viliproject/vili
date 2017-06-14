package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/websocket"

	"github.com/airware/vili/config"
	"github.com/airware/vili/docker"
	"github.com/airware/vili/errors"
	"github.com/airware/vili/kube"
	"github.com/airware/vili/kube/extensions/v1beta1"
	"github.com/airware/vili/log"
	"github.com/airware/vili/server"
	"github.com/airware/vili/session"
	"github.com/airware/vili/templates"
	"github.com/airware/vili/util"
	echo "gopkg.in/labstack/echo.v1"
)

const (
	jobRunTimeout = 5 * time.Minute
)

var (
	jobRunsQueryParams = []string{"labelSelector", "fieldSelector", "resourceVersion"}
)

func jobRunsGetHandler(c *echo.Context) error {
	env := c.Param("env")
	job := c.Param("job")
	query := filterQueryFields(c, jobRunsQueryParams)

	labelSelector := query.Get("labelSelector")
	if labelSelector != "" {
		labelSelector += ","
	}
	labelSelector += "job=" + job
	query.Set("labelSelector", labelSelector)

	if c.Request().URL.Query().Get("watch") != "" {
		// watch jobs and return over websocket
		var err error
		websocket.Handler(func(ws *websocket.Conn) {
			err = jobRunsWatchHandler(ws, env, query)
			ws.Close()
		}).ServeHTTP(c.Response(), c.Request())
		return err
	}

	// otherwise, return the pods list
	resp, _, err := kube.Jobs.List(env, query)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

func jobRunsWatchHandler(ws *websocket.Conn, env string, query *url.Values) error {
	return apiWatchHandler(ws, env, query, kube.Jobs.Watch)
}

func jobRunCreateHandler(c *echo.Context) error {
	env := c.Param("env")
	jobName := c.Param("job")

	jobRun := new(JobRun)
	if err := json.NewDecoder(c.Request().Body).Decode(jobRun); err != nil {
		return err
	}
	if jobRun.Branch == "" {
		return server.ErrorResponse(c, errors.BadRequest("Request missing branch"))
	}
	if jobRun.Tag == "" {
		return server.ErrorResponse(c, errors.BadRequest("Request missing tag"))
	}
	jobRun.Env = env
	jobRun.JobName = jobName
	jobRun.Username = c.Get("user").(*session.User).Username

	err := jobRun.Run(c.Request().URL.Query().Get("async") != "")
	if err != nil {
		switch e := err.(type) {
		case JobRunInitError:
			return server.ErrorResponse(c, errors.BadRequest(e.Error()))
		default:
			return e
		}
	}
	return c.JSON(http.StatusOK, jobRun)
}

// JobRun represents a single pod run
type JobRun struct {
	ID       string    `json:"id"`
	Env      string    `json:"env"`
	JobName  string    `json:"jobName"`
	Branch   string    `json:"branch"`
	Tag      string    `json:"tag"`
	Time     time.Time `json:"time"`
	Username string    `json:"username"`

	Job *v1beta1.Job `json:"job"`
}

// Run initializes a job, checks to make sure it is valid, and runs it
func (r *JobRun) Run(async bool) error {
	r.ID = util.RandLowercaseString(16)
	r.Time = time.Now()

	digest, err := docker.GetTag(r.JobName, r.Branch, r.Tag)
	if err != nil {
		return err
	}
	if digest == "" {
		return JobRunInitError{
			message: fmt.Sprintf("Tag %s not found for job %s", r.Tag, r.JobName),
		}
	}

	err = r.createNewJob()
	if err != nil {
		return err
	}

	if async {
		go r.watchJob()
		return nil
	}
	return r.watchJob()
}

func (r *JobRun) createNewJob() (err error) {
	// get the spec
	jobTemplate, err := templates.Job(r.Env, r.Branch, r.JobName)
	if err != nil {
		return
	}

	job := new(v1beta1.Job)
	err = jobTemplate.Parse(job)
	if err != nil {
		return
	}

	containers := job.Spec.Template.Spec.Containers
	if len(containers) == 0 {
		return fmt.Errorf("no containers in job")
	}

	imageName, err := docker.FullName(r.JobName, r.Branch, r.Tag)
	if err != nil {
		return
	}
	containers[0].Image = imageName

	job.ObjectMeta.Name = r.JobName + "-" + r.ID

	// add labels
	labels := map[string]string{
		"job": r.JobName,
		"run": job.ObjectMeta.Name,
	}
	job.ObjectMeta.Labels = labels
	job.Spec.Template.ObjectMeta.Labels = labels

	// add annotations
	if job.ObjectMeta.Annotations == nil {
		job.ObjectMeta.Annotations = map[string]string{}
	}
	if job.Spec.Template.ObjectMeta.Annotations == nil {
		job.Spec.Template.ObjectMeta.Annotations = map[string]string{}
	}
	job.ObjectMeta.Annotations["vili/branch"] = r.Branch
	job.Spec.Template.ObjectMeta.Annotations["vili/branch"] = r.Branch
	job.ObjectMeta.Annotations["vili/startedBy"] = r.Username
	job.Spec.Template.ObjectMeta.Annotations["vili/startedBy"] = r.Username

	newJob, status, err := kube.Jobs.Create(r.Env, job)
	if err != nil {
		return
	}
	if status != nil {
		return fmt.Errorf(status.Message)
	}
	r.Job = newJob
	r.logMessage(fmt.Sprintf("Job for tag %s and branch %s created by %s", r.Tag, r.Branch, r.Username), log.InfoLevel)
	return
}

// watchJob waits until the job exits
func (r *JobRun) watchJob() (err error) {
	watcher, err := kube.Jobs.Watch(r.Env, &url.Values{
		"fieldSelector": {"metadata.name=" + r.Job.ObjectMeta.Name},
	})
	if err != nil {
		return
	}

	startTime := time.Now()
eventLoop:
	for {
		select {
		case event, ok := <-watcher.EventChan:
			if !ok {
				break eventLoop
			}
			if watcher.Stopped() {
				// empty the channel
				continue
			}
			elapsed := time.Now().Sub(startTime)
			jobEvent := event.(*kube.JobEvent)
			job := jobEvent.Object
			if jobEvent.List != nil && len(jobEvent.List.Items) > 0 {
				job = &jobEvent.List.Items[0]
			}
			switch jobEvent.Type {
			case kube.WatchEventDeleted:
				r.logMessage(fmt.Sprintf("Deleted job after %s", humanizeDuration(elapsed)), log.WarnLevel)
			case kube.WatchEventInit, kube.WatchEventAdded, kube.WatchEventModified:
				finished := false
				for _, condition := range job.Status.Conditions {
					switch condition.Type {
					case v1beta1.JobComplete:
						elapsed := time.Now().Sub(startTime)
						r.logMessage(fmt.Sprintf("Successfully completed job in %s", humanizeDuration(elapsed)), log.InfoLevel)
						finished = true
					case v1beta1.JobFailed:
						elapsed := time.Now().Sub(startTime)
						r.logMessage(fmt.Sprintf("Failed job after %s", humanizeDuration(elapsed)), log.ErrorLevel)
						finished = true
						err = fmt.Errorf("failed")
					}
				}
				if finished {
					watcher.Stop()
					break
				}
			}
		case <-time.After(jobRunTimeout):
			elapsed := time.Now().Sub(startTime)
			r.logMessage(fmt.Sprintf("Job timed out after %s", humanizeDuration(elapsed)), log.WarnLevel)
			watcher.Stop()
			err = fmt.Errorf("timeout")
			break
		}
	}
	return
}

func (r *JobRun) logMessage(message string, level log.Level) {
	urlStr := fmt.Sprintf(
		"%s/%s/jobs/%s/runs/%s",
		config.GetString(config.URI),
		r.Env,
		r.JobName,
		r.Job.ObjectMeta.Name,
	)
	slackMessage := fmt.Sprintf(
		"*%s* - *%s* - <%s|%s> - %s",
		r.Env,
		r.JobName,
		urlStr,
		r.ID,
		message,
	)
	jobMessage := fmt.Sprintf(
		"%s - %s - %s",
		r.Env,
		r.JobName,
		message,
	)
	logMessage(jobMessage, slackMessage, level)
}

// JobRunInitError is raised if there is a problem initializing a pod
type JobRunInitError struct {
	message string
}

func (e JobRunInitError) Error() string {
	return e.message
}
