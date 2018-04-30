package api

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/airware/vili/log"
	"github.com/labstack/echo"
	"golang.org/x/net/websocket"
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

func getImageTagFromDeployment(deployment *extv1beta1.Deployment) (string, error) {
	containers := deployment.Spec.Template.Spec.Containers
	if len(containers) == 0 {
		return "", fmt.Errorf("no containers in deployment")
	}
	image := containers[0].Image
	imageSplit := strings.Split(image, ":")
	if len(imageSplit) != 2 {
		return "", fmt.Errorf("invalid image: %s", image)
	}
	return imageSplit[1], nil
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
