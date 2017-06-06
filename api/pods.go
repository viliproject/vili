package api

import (
	"io"
	"net/http"
	"net/url"

	"golang.org/x/net/websocket"

	"github.com/airware/vili/kube"
	"github.com/airware/vili/log"
	echo "gopkg.in/labstack/echo.v1"
)

var (
	podsQueryParams   = []string{"labelSelector", "fieldSelector", "resourceVersion"}
	podLogQueryParams = []string{"sinceSeconds", "sinceTime"}
)

func podsHandler(c *echo.Context) error {
	env := c.Param("env")
	query := filterQueryFields(c, podsQueryParams)

	if c.Request().URL.Query().Get("watch") != "" {
		// watch pods and return over websocket
		var err error
		websocket.Handler(func(ws *websocket.Conn) {
			err = podsWatchHandler(ws, env, query)
			ws.Close()
		}).ServeHTTP(c.Response(), c.Request())
		return err
	}

	// otherwise, return the pods list
	resp, _, err := kube.Pods.List(env, query)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

func podsWatchHandler(ws *websocket.Conn, env string, query *url.Values) error {
	return apiWatchHandler(ws, env, query, kube.Pods.Watch)
}

func podLogHandler(c *echo.Context) error {
	env := c.Param("env")
	name := c.Param("pod")
	query := filterQueryFields(c, podLogQueryParams)

	if c.Request().URL.Query().Get("follow") != "" {
		// watch pod logs and return changes over websocket
		var err error
		websocket.Handler(func(ws *websocket.Conn) {
			err = podLogWatchHandler(ws, env, name, query)
			ws.Close()
		}).ServeHTTP(c.Response(), c.Request())
		return err
	}

	resp, status, err := kube.Pods.GetLog(env, name)
	if err != nil {
		return err
	}
	if status != nil {
		return c.JSON(http.StatusOK, status)
	}
	return c.JSON(http.StatusOK, resp)
}

func podLogWatchHandler(ws *websocket.Conn, env, name string, query *url.Values) error {
	watcher, err := kube.Pods.WatchLog(env, name, query)
	if err != nil {
		return err
	}

	go func() {
		var cmd interface{}
		err := websocket.JSON.Receive(ws, cmd)
		if err == io.EOF {
			watcher.Stop()
		}
	}()

	first := true
	for logLine := range watcher.EventChan {
		logType := "ADD"
		if first {
			logType = "START"
			first = false
		}
		err := websocket.JSON.Send(ws, map[string]string{
			"type":   logType,
			"object": logLine.(string),
		})
		if err != nil {
			log.WithError(err).Error("error writing to pod log stream")
			watcher.Stop()
		}
	}

	if watcher.Err() != nil {
		return watcher.Err()
	}
	if !watcher.Stopped() {
		websocket.JSON.Send(ws, webSocketCloseMessage)
	}
	return nil
}

func podDeleteHandler(c *echo.Context) error {
	env := c.Param("env")
	pod := c.Param("pod")

	resp, _, err := kube.Pods.Delete(env, pod)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}
