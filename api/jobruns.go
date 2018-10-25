package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/viliproject/vili/config"
	"github.com/viliproject/vili/errors"
	"github.com/viliproject/vili/kube"
	"github.com/viliproject/vili/log"
	"github.com/viliproject/vili/repository"
	"github.com/viliproject/vili/server"
	"github.com/viliproject/vili/session"
	"github.com/viliproject/vili/templates"
	"github.com/viliproject/vili/util"
	"github.com/labstack/echo"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

func jobRunsGetHandler(c echo.Context) error {
	env := c.Param("env")
	job := c.Param("job")

	endpoint := kube.GetClient(env).Jobs()
	query := getListOptionsFromRequest(c)
	if query.LabelSelector != "" {
		query.LabelSelector += ","
	}
	query.LabelSelector += "job=" + job

	if c.Request().URL.Query().Get("watch") != "" {
		return apiWatchWebsocket(c, query, endpoint.Watch)
	}

	// otherwise, return the pods list
	resp, err := endpoint.List(query)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

func jobRunCreateHandler(c echo.Context) error {
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

	Job *batchv1.Job `json:"job"`
}

// Run initializes a job, checks to make sure it is valid, and runs it
func (r *JobRun) Run(async bool) error {
	r.ID = util.RandLowercaseString(16)
	r.Time = time.Now()

	digest, err := repository.GetDockerTag(r.JobName, r.Tag)
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

	job := new(batchv1.Job)
	err = jobTemplate.Parse(job)
	if err != nil {
		return
	}

	containers := job.Spec.Template.Spec.Containers
	if len(containers) == 0 {
		return fmt.Errorf("no containers in job")
	}

	imageName, err := repository.DockerFullName(r.JobName, r.Tag)
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

	newJob, err := kube.GetClient(r.Env).Jobs().Create(job)
	if err != nil {
		return
	}
	r.Job = newJob
	r.logMessage(fmt.Sprintf("Job for tag %s and branch %s created by %s", r.Tag, r.Branch, r.Username), log.InfoLevel)
	return
}

// watchJob waits until the job exits
func (r *JobRun) watchJob() (err error) {
	watcher, err := kube.GetClient(r.Env).Jobs().Watch(metav1.ListOptions{
		FieldSelector: "metadata.name=" + r.Job.ObjectMeta.Name,
	})
	if err != nil {
		return
	}

	startTime := time.Now()
eventLoop:
	for {
		select {
		case event, ok := <-watcher.ResultChan():
			if !ok {
				break eventLoop
			}
			elapsed := time.Now().Sub(startTime)
			job := event.Object.(*batchv1.Job)
			switch event.Type {
			case watch.Deleted:
				r.logMessage(fmt.Sprintf("Deleted job after %s", humanizeDuration(elapsed)), log.WarnLevel)
				watcher.Stop()
				err = fmt.Errorf("deleted")
				break eventLoop
			case watch.Added, watch.Modified:
				finished := false
				for _, condition := range job.Status.Conditions {
					switch condition.Type {
					case batchv1.JobComplete:
						elapsed := time.Now().Sub(startTime)
						r.logMessage(fmt.Sprintf("Successfully completed job in %s", humanizeDuration(elapsed)), log.InfoLevel)
						finished = true
					case batchv1.JobFailed:
						elapsed := time.Now().Sub(startTime)
						r.logMessage(fmt.Sprintf("Failed job after %s", humanizeDuration(elapsed)), log.ErrorLevel)
						finished = true
						err = fmt.Errorf("failed")
					}
				}
				if finished {
					watcher.Stop()
					break eventLoop
				}
			}
		case <-time.After(config.GetDuration(config.JobRunTimeout)):
			elapsed := time.Now().Sub(startTime)
			r.logMessage(fmt.Sprintf("Job timed out after %s", humanizeDuration(elapsed)), log.WarnLevel)
			watcher.Stop()
			err = fmt.Errorf("timeout")
			break eventLoop
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
