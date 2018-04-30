package api

import (
	"bufio"
	"io"
	"net/http"
	"strconv"

	"github.com/airware/vili/kube"
	"github.com/airware/vili/log"
	"github.com/labstack/echo"
	"golang.org/x/net/websocket"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
)

var (
	podLogQueryParams = []string{"sinceSeconds", "sinceTime"}
)

func podsHandler(c echo.Context) error {
	env := c.Param("env")

	endpoint := kube.GetClient(env).Pods()
	query := getListOptionsFromRequest(c)

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

func podLogHandler(c echo.Context) error {
	env := c.Param("env")
	name := c.Param("pod")

	endpoint := kube.GetClient(env).Pods()
	query := parsePodLogOptions(c)
	logRequest := endpoint.GetLogs(name, query)

	if query.Follow {
		// watch pod logs and return changes over websocket
		var err error
		websocket.Handler(func(ws *websocket.Conn) {
			err = podLogWatchHandler(ws, logRequest)
			ws.Close()
		}).ServeHTTP(c.Response(), c.Request())
		return err
	}

	output, err := logRequest.DoRaw()
	if err != nil {
		return err
	}
	return c.String(http.StatusOK, string(output))
}

func parsePodLogOptions(c echo.Context) *corev1.PodLogOptions {
	urlQuery := c.Request().URL.Query()
	query := &corev1.PodLogOptions{}
	query.Follow, _ = strconv.ParseBool(urlQuery.Get("follow"))
	return query
}

func podLogWatchHandler(ws *websocket.Conn, request *rest.Request) error {
	readCloser, err := request.Stream()
	if err != nil {
		return err
	}

	go func() {
		var cmd interface{}
		err := websocket.JSON.Receive(ws, cmd)
		if err == io.EOF {
			readCloser.Close()
		}
	}()

	go func() {
		first := true
		scanner := bufio.NewScanner(readCloser)
		for scanner.Scan() {
			err := scanner.Err()
			if err != nil {
				log.WithError(err).Warn("error scanning")
				break
			}
			logType := "ADD"
			if first {
				logType = "START"
				first = false
			}
			err = websocket.JSON.Send(ws, map[string]string{
				"type":   logType,
				"object": scanner.Text(),
			})
			if err != nil {
				log.WithError(err).Error("error writing to pod log stream")
				readCloser.Close()
			}
		}
	}()

	<-ExitingChan
	readCloser.Close()
	websocket.JSON.Send(ws, webSocketCloseMessage)
	return nil
}

func podDeleteHandler(c echo.Context) error {
	env := c.Param("env")
	pod := c.Param("pod")

	endpoint := kube.GetClient(env).Pods()

	err := endpoint.Delete(pod, nil)
	if err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}
