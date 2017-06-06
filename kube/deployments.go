package kube

import (
	"bytes"
	"encoding/json"
	"net/url"

	"github.com/airware/vili/kube/extensions/v1beta1"
	"github.com/airware/vili/kube/unversioned"
	"github.com/airware/vili/log"
)

// Deployments is the default deployments service instance
var Deployments = new(DeploymentsService)

// DeploymentsService is the kubernetes service to interace with deployments
type DeploymentsService struct {
}

// List fetches the list of deployments in `env`
func (s *DeploymentsService) List(env string, query *url.Values) (*v1beta1.DeploymentList, *unversioned.Status, error) {
	client, err := getClient(env)
	if err != nil {
		return nil, nil, invalidEnvError(env)
	}
	resp := new(v1beta1.DeploymentList)
	status, err := client.unmarshalRequest("GET", "deployments", query, nil, resp)
	if status != nil || err != nil {
		return nil, status, err
	}
	return resp, nil, nil
}

// Get fetches the deployment in `env` with `name`
func (s *DeploymentsService) Get(env, name string) (*v1beta1.Deployment, *unversioned.Status, error) {
	client, err := getClient(env)
	if err != nil {
		return nil, nil, invalidEnvError(env)
	}
	resp := new(v1beta1.Deployment)
	status, err := client.unmarshalRequest("GET", "deployments/"+name, nil, nil, resp)
	if status != nil || err != nil {
		return nil, status, err
	}
	return resp, nil, nil
}

// Create creates a deployment in `env`
func (s *DeploymentsService) Create(env string, data *v1beta1.Deployment) (*v1beta1.Deployment, *unversioned.Status, error) {
	client, err := getClient(env)
	if err != nil {
		return nil, nil, invalidEnvError(env)
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, nil, err
	}
	resp := new(v1beta1.Deployment)
	status, err := client.unmarshalRequest(
		"POST",
		"deployments",
		nil,
		bytes.NewReader(dataBytes),
		resp,
	)
	if status != nil || err != nil {
		return nil, status, err
	}
	return resp, nil, nil
}

// Replace replaces the deployment in `env` with `name`
func (s *DeploymentsService) Replace(env, name string, data *v1beta1.Deployment) (*v1beta1.Deployment, *unversioned.Status, error) {
	client, err := getClient(env)
	if err != nil {
		return nil, nil, invalidEnvError(env)
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, nil, err
	}
	resp := new(v1beta1.Deployment)
	status, err := client.unmarshalRequest(
		"PUT",
		"deployments/"+name,
		nil,
		bytes.NewReader(dataBytes),
		resp,
	)
	if status != nil || err != nil {
		return nil, status, err
	}
	return resp, nil, nil
}

// Scale scales the deployment in `env` with `name`
func (s *DeploymentsService) Scale(env, name string, data *v1beta1.Scale) (*v1beta1.Scale, *unversioned.Status, error) {
	client, err := getClient(env)
	if err != nil {
		return nil, nil, invalidEnvError(env)
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, nil, err
	}
	resp := new(v1beta1.Scale)
	status, err := client.unmarshalRequest(
		"PATCH",
		"deployments/"+name+"/scale",
		nil,
		bytes.NewReader(dataBytes),
		resp,
	)
	if status != nil || err != nil {
		return nil, status, err
	}
	return resp, nil, nil
}

// Rollback rolls back the deployment in `env` with `name`
func (s *DeploymentsService) Rollback(env, name string, data *v1beta1.DeploymentRollback) (*v1beta1.DeploymentRollback, *unversioned.Status, error) {
	client, err := getClient(env)
	if err != nil {
		return nil, nil, invalidEnvError(env)
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, nil, err
	}
	log.Warn(string(dataBytes))
	resp := new(v1beta1.DeploymentRollback)
	status, err := client.unmarshalRequest(
		"POST",
		"deployments/"+name+"/rollback",
		nil,
		bytes.NewReader(dataBytes),
		resp,
	)
	if status != nil || err != nil {
		return nil, status, err
	}
	return resp, nil, nil
}

// Delete deletes the deployment in `env` with `name`
func (s *DeploymentsService) Delete(env, name string) (*unversioned.Status, error) {
	client, err := getClient(env)
	if err != nil {
		return nil, invalidEnvError(env)
	}
	status, err := client.unmarshalRequest("DELETE", "deployments/"+name, nil, nil, nil)
	if status != nil || err != nil {
		return status, err
	}
	return nil, nil
}

// DeploymentEvent describes an event on a deployment
type DeploymentEvent struct {
	Type   WatchEventType          `json:"type"`
	Object *v1beta1.Deployment     `json:"object"`
	List   *v1beta1.DeploymentList `json:"list"`
}

// Watch watches deployments in `env`
func (s *DeploymentsService) Watch(env string, query *url.Values) (watcher *Watcher, err error) {
	return watchObjectStream(env, "deployments", query, func(eventType WatchEventType, body json.RawMessage) (interface{}, error) {
		if eventType == WatchEventInit {
			event := &DeploymentEvent{
				Type: eventType,
				List: new(v1beta1.DeploymentList),
			}
			return event, json.Unmarshal(body, event.List)
		}
		event := &DeploymentEvent{
			Type:   eventType,
			Object: new(v1beta1.Deployment),
		}
		return event, json.Unmarshal(body, event.Object)
	})
}
