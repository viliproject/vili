package api

import (
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/websocket"

	"github.com/airware/vili/kube"
	"github.com/airware/vili/kube/extensions/v1beta1"
	"github.com/airware/vili/log"
	echo "gopkg.in/labstack/echo.v1"
)

var (
	webSocketCloseMessage = map[string]string{
		"type": "CLOSED",
	}
)

func parseQueryFields(c *echo.Context) map[string]bool {
	queryFields := make(map[string]bool)
	requestFields := c.Request().URL.Query().Get("fields")
	if requestFields != "" {
		for _, field := range strings.Split(requestFields, ",") {
			queryFields[field] = true
		}
	}
	return queryFields
}

func filterQueryFields(c *echo.Context, params []string) *url.Values {
	query := &url.Values{}
	for _, param := range params {
		val := c.Request().URL.Query().Get(param)
		if val != "" {
			query.Add(param, val)
		}
	}
	return query
}

func getPortFromDeployment(deployment *v1beta1.Deployment) (int32, error) {
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

func getImageTagFromDeployment(deployment *v1beta1.Deployment) (string, error) {
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

func apiWatchHandler(ws *websocket.Conn, env string, query *url.Values, watchFunc func(env string, query *url.Values) (*kube.Watcher, error)) error {
	watcher, err := watchFunc(env, query)
	if err != nil {
		return err
	}

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

	for event := range watcher.EventChan {
		err := websocket.JSON.Send(ws, event)
		if err != nil {
			log.WithError(err).Warn("error writing to websocket stream")
			watcher.Stop()
		}
	}
	if !watcher.Stopped() {
		watcher.Stop()
		websocket.JSON.Send(ws, webSocketCloseMessage)
	}
	return watcher.Err()
}
