package kube

import (
	"bytes"
	"encoding/json"
	"net/url"

	"github.com/airware/vili/kube/unversioned"
	"github.com/airware/vili/kube/v1"
)

// ConfigMaps is the default configmaps service instance
var ConfigMaps = new(ConfigMapsService)

// ConfigMapsService is the kubernetes service to interace with configmaps
type ConfigMapsService struct {
}

// List fetches the list of configmaps in `env`
func (s *ConfigMapsService) List(env string, query *url.Values) (*v1.ConfigMapList, *unversioned.Status, error) {
	client, err := getClient(env)
	if err != nil {
		return nil, nil, invalidEnvError(env)
	}
	resp := new(v1.ConfigMapList)
	status, err := client.unmarshalRequest("GET", "configmaps", query, nil, resp)
	if status != nil || err != nil {
		return nil, status, err
	}
	return resp, nil, nil
}

// Get fetches the configmap in `env` with `name`
func (s *ConfigMapsService) Get(env, name string) (*v1.ConfigMap, *unversioned.Status, error) {
	client, err := getClient(env)
	if err != nil {
		return nil, nil, invalidEnvError(env)
	}
	resp := new(v1.ConfigMap)
	status, err := client.unmarshalRequest("GET", "configmaps/"+name, nil, nil, resp)
	if status != nil || err != nil {
		return nil, status, err
	}
	return resp, nil, nil
}

// Create creates a configmap in `env`
func (s *ConfigMapsService) Create(env string, data *v1.ConfigMap) (*v1.ConfigMap, *unversioned.Status, error) {
	client, err := getClient(env)
	if err != nil {
		return nil, nil, invalidEnvError(env)
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, nil, err
	}
	resp := new(v1.ConfigMap)
	status, err := client.unmarshalRequest(
		"POST",
		"configmaps",
		nil,
		bytes.NewReader(dataBytes),
		resp,
	)
	if status != nil || err != nil {
		return nil, status, err
	}
	return resp, nil, nil
}

// Replace replaces the configmap in `env` with `name`
func (s *ConfigMapsService) Replace(env, name string, data *v1.ConfigMap) (*v1.ConfigMap, *unversioned.Status, error) {
	client, err := getClient(env)
	if err != nil {
		return nil, nil, invalidEnvError(env)
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, nil, err
	}
	resp := new(v1.ConfigMap)
	status, err := client.unmarshalRequest(
		"PUT",
		"configmaps/"+name,
		nil,
		bytes.NewReader(dataBytes),
		resp,
	)
	if status != nil || err != nil {
		return nil, status, err
	}
	return resp, nil, nil
}

// Delete deletes the configmap in `env` with `name`
func (s *ConfigMapsService) Delete(env, name string) (*unversioned.Status, error) {
	client, err := getClient(env)
	if err != nil {
		return nil, invalidEnvError(env)
	}
	status, err := client.unmarshalRequest("DELETE", "configmaps/"+name, nil, nil, nil)
	if status != nil || err != nil {
		return status, err
	}
	return nil, nil
}

// ConfigMapEvent describes an event on a configmap
type ConfigMapEvent struct {
	Type   WatchEventType    `json:"type"`
	Object *v1.ConfigMap     `json:"object"`
	List   *v1.ConfigMapList `json:"list"`
}

// Watch watches configMaps in `env`
func (s *ConfigMapsService) Watch(env string, query *url.Values) (watcher *Watcher, err error) {
	return watchObjectStream(env, "configmaps", query, func(eventType WatchEventType, body json.RawMessage) (interface{}, error) {
		if eventType == WatchEventInit {
			event := &ConfigMapEvent{
				Type: eventType,
				List: new(v1.ConfigMapList),
			}
			return event, json.Unmarshal(body, event.List)
		}
		event := &ConfigMapEvent{
			Type:   eventType,
			Object: new(v1.ConfigMap),
		}
		return event, json.Unmarshal(body, event.Object)
	})
}
