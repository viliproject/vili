package api

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/airware/vili/functions"
	"github.com/airware/vili/log"
	"github.com/airware/vili/templates"
	"github.com/labstack/echo"
	"golang.org/x/net/websocket"
	batchv1 "k8s.io/api/batch/v1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
)

var (
	webSocketCloseMessage = map[string]string{
		"type": "CLOSED",
	}
)

func parseQueryFields(c echo.Context) map[string]bool {
	queryFields := make(map[string]bool)
	requestFields := c.Request().URL.Query().Get("fields")
	if requestFields != "" {
		for _, field := range strings.Split(requestFields, ",") {
			queryFields[field] = true
		}
	}
	return queryFields
}

func getListOptionsFromRequest(c echo.Context) metav1.ListOptions {
	return metav1.ListOptions{
		LabelSelector:   c.Request().URL.Query().Get("labelSelector"),
		FieldSelector:   c.Request().URL.Query().Get("fieldSelector"),
		ResourceVersion: c.Request().URL.Query().Get("resourceVersion"),
	}
}

func getListOptionsForDeployment(deployment *extv1beta1.Deployment) metav1.ListOptions {
	var selector []string
	for k, v := range deployment.Spec.Selector.MatchLabels {
		selector = append(selector, fmt.Sprintf("%s=%s", k, v))
	}
	return metav1.ListOptions{
		LabelSelector: strings.Join(selector, ","),
	}
}

func getDeploymentWithTag(deploymentName, env, branch, tag string) (*extv1beta1.Deployment, error) {
	deploymentTemplate, err := templates.Deployment(env, branch, deploymentName)
	if err != nil {
		return nil, err
	}
	deploymentTemplate, err = deploymentTemplate.Populate(map[string]string{
		"Tag":       tag,
		"Namespace": "",
	})
	if err != nil {
		return nil, err
	}

	deployment := new(extv1beta1.Deployment)
	err = deploymentTemplate.Parse(deployment)
	if err != nil {
		return nil, err
	}
	return deployment, nil
}

func getImageRepoFromDeployment(deployment *extv1beta1.Deployment) (string, error) {
	containers := deployment.Spec.Template.Spec.Containers
	if len(containers) == 0 {
		return "", fmt.Errorf("no containers in controller")
	}
	image := deployment.Spec.Template.Spec.Containers[0].Image
	imageSplit := strings.Split(image, ":")
	if len(imageSplit) != 2 {
		return "", fmt.Errorf("invalid image: %s", image)
	}
	return imageSplit[0], nil
}

func getPortFromDeployment(deployment *extv1beta1.Deployment) (int32, error) {
	containers := deployment.Spec.Template.Spec.Containers
	if len(containers) == 0 {
		return 0, fmt.Errorf("no containers in controller")
	}
	ports := containers[0].Ports
	if len(ports) == 0 {
		return 0, fmt.Errorf("no ports in controller")
	}
	return ports[0].ContainerPort, nil
}

func getDeploymentImageRepo(deploymentName, env, branch string) (string, error) {
	deployment, err := getDeploymentWithTag(deploymentName, env, branch, "tag")
	if err != nil {
		return "", err
	}
	return getImageRepoFromDeployment(deployment)
}

func getJobWithTag(jobName, env, branch, tag string) (*batchv1.Job, error) {
	jobTemplate, err := templates.Job(env, branch, jobName)
	if err != nil {
		return nil, err
	}
	jobTemplate, err = jobTemplate.Populate(map[string]string{
		"Tag":       tag,
		"Namespace": "",
	})
	if err != nil {
		return nil, err
	}

	job := new(batchv1.Job)
	err = jobTemplate.Parse(job)
	if err != nil {
		return nil, err
	}
	return job, nil
}

func getImageRepoFromJob(job *batchv1.Job) (string, error) {
	containers := job.Spec.Template.Spec.Containers
	if len(containers) == 0 {
		return "", fmt.Errorf("no containers in controller")
	}
	image := job.Spec.Template.Spec.Containers[0].Image
	imageSplit := strings.Split(image, ":")
	if len(imageSplit) != 2 {
		return "", fmt.Errorf("invalid image: %s", image)
	}
	return imageSplit[0], nil
}

func getJobImageRepo(jobName, env, branch string) (string, error) {
	job, err := getJobWithTag(jobName, env, branch, "tag")
	if err != nil {
		return "", err
	}
	return getImageRepoFromJob(job)
}

func getFunctionCodeRepo(functionName, env, branch string) (string, error) {
	functionTemplate, err := templates.Function(env, branch, functionName)
	if err != nil {
		return "", err
	}
	functionTemplate, err = functionTemplate.Populate(map[string]string{
		"Tag":              "tag",
		"Namespace":        "",
		"AWSAccountNumber": "",
	})
	if err != nil {
		return "", err
	}

	function := new(functions.FunctionSpec)
	err = functionTemplate.Parse(function)
	if err != nil {
		return "", err
	}
	codeSplit := strings.Split(function.Code, "/")
	if len(codeSplit) < 1 {
		return "", fmt.Errorf("invalid function code: %s", function.Code)
	}
	return strings.Join(codeSplit[0:len(codeSplit)-1], "/"), nil
}

func humanizeDuration(d time.Duration) string {
	return ((d / time.Second) * time.Second).String()
}

type apiWatcher func(opts metav1.ListOptions) (watch.Interface, error)

// apiEvent is used for json serialization of kubernetes watch events
type apiEvent struct {
	Type   watch.EventType `json:"type"`
	Object runtime.Object  `json:"object"`
}

func apiWatchWebsocket(c echo.Context, query metav1.ListOptions, watchFunc apiWatcher) (err error) {
	websocket.Handler(func(ws *websocket.Conn) {
		err = apiWatchHandler(ws, query, watchFunc)
		ws.Close()
	}).ServeHTTP(c.Response(), c.Request())
	return
}

func apiWatchHandler(ws *websocket.Conn, query metav1.ListOptions, watchFunc apiWatcher) error {
	watcher, err := watchFunc(query)
	if err != nil {
		return err
	}
	stoppedChan := make(chan struct{})

	go func() {
		var cmd interface{}
		for {
			err := websocket.JSON.Receive(ws, cmd)
			if err == io.EOF {
				watcher.Stop()
				break
			}
		}
	}()

	go func() {
		for event := range watcher.ResultChan() {
			err := websocket.JSON.Send(ws, apiEvent(event))
			if err != nil {
				log.WithError(err).Warn("error writing to websocket stream")
				watcher.Stop()
			}
		}
		close(stoppedChan)
	}()

	select {
	case <-ExitingChan:
		watcher.Stop()
	case <-stoppedChan:
	}
	websocket.JSON.Send(ws, webSocketCloseMessage)
	return nil
}
