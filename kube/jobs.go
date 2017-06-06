package kube

import (
	"bytes"
	"encoding/json"
	"net/url"

	"github.com/airware/vili/kube/extensions/v1beta1"
	"github.com/airware/vili/kube/unversioned"
)

// Jobs is the default jobs service instance
var Jobs = new(JobsService)

// JobsService is the kubernetes service to interace with jobs
type JobsService struct {
}

// List fetches the list of jobs in `env`
func (s *JobsService) List(env string, query *url.Values) (*v1beta1.JobList, *unversioned.Status, error) {
	client, err := getClient(env)
	if err != nil {
		return nil, nil, invalidEnvError(env)
	}
	resp := new(v1beta1.JobList)
	status, err := client.unmarshalRequest("GET", "jobs", query, nil, resp)
	if status != nil || err != nil {
		return nil, status, err
	}
	return resp, nil, nil
}

// Get fetches the job in `env` with `name`
func (s *JobsService) Get(env, name string) (*v1beta1.Job, *unversioned.Status, error) {
	client, err := getClient(env)
	if err != nil {
		return nil, nil, invalidEnvError(env)
	}
	resp := new(v1beta1.Job)
	status, err := client.unmarshalRequest("GET", "jobs/"+name, nil, nil, resp)
	if status != nil || err != nil {
		return nil, status, err
	}
	return resp, nil, nil
}

// Create creates a job in `env`
func (s *JobsService) Create(env string, data *v1beta1.Job) (*v1beta1.Job, *unversioned.Status, error) {
	client, err := getClient(env)
	if err != nil {
		return nil, nil, invalidEnvError(env)
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, nil, err
	}
	resp := new(v1beta1.Job)
	status, err := client.unmarshalRequest(
		"POST",
		"jobs",
		nil,
		bytes.NewReader(dataBytes),
		resp,
	)
	if status != nil || err != nil {
		return nil, status, err
	}
	return resp, nil, nil
}

// Delete deletes the job in `env` with `name`
func (s *JobsService) Delete(env, name string) (*unversioned.Status, error) {
	client, err := getClient(env)
	if err != nil {
		return nil, invalidEnvError(env)
	}
	return client.unmarshalRequest("DELETE", "jobs/"+name, nil, nil, nil)
}

// JobEvent describes an event on a job
type JobEvent struct {
	Type   WatchEventType   `json:"type"`
	Object *v1beta1.Job     `json:"object"`
	List   *v1beta1.JobList `json:"list"`
}

// Watch watches jobs in `env`
func (s *JobsService) Watch(env string, query *url.Values) (watcher *Watcher, err error) {
	return watchObjectStream(env, "jobs", query, func(eventType WatchEventType, body json.RawMessage) (interface{}, error) {
		if eventType == WatchEventInit {
			event := &JobEvent{
				Type: eventType,
				List: new(v1beta1.JobList),
			}
			return event, json.Unmarshal(body, event.List)
		}
		event := &JobEvent{
			Type:   eventType,
			Object: new(v1beta1.Job),
		}
		return event, json.Unmarshal(body, event.Object)
	})
}
