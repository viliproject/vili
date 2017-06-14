package kube

import (
	"bufio"
	"encoding/json"
	"net/url"
	"sync"

	"github.com/airware/vili/errors"
	"github.com/airware/vili/kube/unversioned"
	"github.com/airware/vili/log"
)

// WatchEventType describes an event type
type WatchEventType string

// Possible values of WatchEventType
const (
	WatchEventInit     WatchEventType = "INIT"
	WatchEventAdded    WatchEventType = "ADDED"
	WatchEventModified WatchEventType = "MODIFIED"
	WatchEventDeleted  WatchEventType = "DELETED"
	WatchEventError    WatchEventType = "ERROR"
)

// WatchEvent describes an event
type WatchEvent struct {
	Type   WatchEventType  `json:"type"`
	Object json.RawMessage `json:"object"`
}

// Watcher is a struct that returns the events channel and the ability to stop watching
type Watcher struct {
	EventChan chan interface{}
	stopChan  chan struct{}
	err       error
	stopped   bool
	mutex     sync.RWMutex
}

func newWatcher() *Watcher {
	return &Watcher{
		EventChan: make(chan interface{}),
		stopChan:  make(chan struct{}),
	}
}

// sendEvent sends an event to the event channel
func (w *Watcher) sendEvent(event interface{}) {
	if !w.stopped {
		w.EventChan <- event
	}
}

// Stop stops the watcher
func (w *Watcher) Stop() {
	w.mutex.Lock()
	if !w.stopped {
		close(w.stopChan)
		w.stopped = true
	}
	w.mutex.Unlock()
}

// Stopped returns whether this watcher is stopped
func (w *Watcher) Stopped() bool {
	w.mutex.RLock()
	defer w.mutex.RUnlock()
	return w.stopped
}

// Err returns the error for this watcher
func (w *Watcher) Err() error {
	return w.err
}

func (c *client) streamWatchRequest(path string, query *url.Values, watcher *Watcher, makeEvent func([]byte) (interface{}, error)) {
	// create request
	req, err := c.createRequest("GET", path, query, nil, true)
	if err != nil {
		watcher.err = err
		return
	}

	// send request
	resp, err := c.httpClientNoTimeout.Do(req)
	if err != nil {
		watcher.err = err
		return
	}

	scanner := bufio.NewScanner(resp.Body)

	// watch both the stopChan and ExitingChan channels and close the response body if they are triggered
	go func() {
		select {
		case <-watcher.stopChan:
			break
		case <-ExitingChan:
			break
		}

		if err := resp.Body.Close(); err != nil {
			log.WithError(err).Error("error closing response body")
		}
	}()

	// scan the response body for new events
	go func() {
		for scanner.Scan() {
			err := scanner.Err()
			if err != nil {
				log.WithError(err).Warn("error scanning")
				break
			}
			event, err := makeEvent(scanner.Bytes())
			if err != nil {
				log.WithError(err).Warn("error processing watch response")
				break
			}
			watcher.sendEvent(event)
		}
		close(watcher.EventChan)
	}()
}

func watchObjectStream(env, path string, query *url.Values, makeEvent func(WatchEventType, json.RawMessage) (interface{}, error)) (watcher *Watcher, err error) {
	// get env client
	client, err := getClient(env)
	if err != nil {
		err = invalidEnvError(env)
		return
	}
	// initialize query if not initialized
	if query == nil {
		query = &url.Values{}
	}

	var firstEvent interface{}
	// get list of objects first if resourceVersion is not provided
	if query.Get("resourceVersion") == "" {
		respBody, status, err := client.getRequestBytes("GET", path, query, nil)
		if err != nil {
			return nil, err
		}
		if status != nil {
			return nil, errors.BadRequest(status.Message)
		}
		obj := new(kubeObject)
		err = json.Unmarshal(respBody, obj)
		if err != nil {
			return nil, err
		}
		event, err := makeEvent(WatchEventInit, respBody)
		if err != nil {
			return nil, err
		}
		query.Set("resourceVersion", obj.Metadata.ResourceVersion)
		firstEvent = event
	}

	// create new watcher here, so that errors returned from above don't leave a watcher hanging
	watcher = newWatcher()

	if firstEvent != nil {
		go func() {
			watcher.sendEvent(firstEvent)
		}()
	}

	log.Debugf("subscribing to %s events - %s", path, env)
	client.streamWatchRequest(path, query, watcher, func(b []byte) (interface{}, error) {
		event := new(WatchEvent)
		err := json.Unmarshal(b, event)
		if err != nil {
			log.WithError(err).Error("error parsing watch response")
			return nil, err
		}
		if event.Type == "" || len(event.Object) == 0 {
			return nil, nil
		} else if event.Type == WatchEventError {
			status := new(unversioned.Status)
			err = json.Unmarshal(event.Object, status)
			if err == nil {
				err = errors.BadRequest(status.Message)
			}
			log.WithError(err).Error("error event from watch response")
			return nil, err
		}
		return makeEvent(event.Type, event.Object)
	})
	return watcher, nil
}

type kubeObject struct {
	Metadata struct {
		ResourceVersion string `json:"resourceVersion,omitempty"`
	} `json:"metadata,omitempty"`
}
