package kube

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/airware/vili/kube/extensions/v1beta1"
	"github.com/airware/vili/kube/unversioned"
)

// ReplicaSets is the default replicasets service instance
var ReplicaSets = new(ReplicaSetsService)

// ReplicaSetsService is the kubernetes service to interace with replicasets
type ReplicaSetsService struct {
}

// List fetches the list of replicasets in `env`
func (s *ReplicaSetsService) List(env string, query *url.Values) (*v1beta1.ReplicaSetList, *unversioned.Status, error) {
	client, err := getClient(env)
	if err != nil {
		return nil, nil, invalidEnvError(env)
	}
	resp := new(v1beta1.ReplicaSetList)
	status, err := client.unmarshalRequest("GET", "replicasets", query, nil, resp)
	if status != nil || err != nil {
		return nil, status, err
	}
	return resp, nil, nil
}

// ListForDeployment fetches the list of replicasets in `env` for the given deployment
func (s *ReplicaSetsService) ListForDeployment(env string, deployment *v1beta1.Deployment) (*v1beta1.ReplicaSetList, *unversioned.Status, error) {
	var selector []string
	for k, v := range deployment.Spec.Selector.MatchLabels {
		selector = append(selector, fmt.Sprintf("%s=%s", k, v))
	}
	query := &url.Values{
		"labelSelector": {strings.Join(selector, ",")},
	}
	return s.List(env, query)
}

// Get fetches the replicaset in `env` with `name`
func (s *ReplicaSetsService) Get(env, name string) (*v1beta1.ReplicaSet, *unversioned.Status, error) {
	client, err := getClient(env)
	if err != nil {
		return nil, nil, invalidEnvError(env)
	}
	resp := new(v1beta1.ReplicaSet)
	status, err := client.unmarshalRequest("GET", "replicasets/"+name, nil, nil, resp)
	if status != nil || err != nil {
		return nil, status, err
	}
	return resp, nil, nil
}

// Create creates a replicaset in `env`
func (s *ReplicaSetsService) Create(env string, data *v1beta1.ReplicaSet) (*v1beta1.ReplicaSet, *unversioned.Status, error) {
	client, err := getClient(env)
	if err != nil {
		return nil, nil, invalidEnvError(env)
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, nil, err
	}
	resp := new(v1beta1.ReplicaSet)
	status, err := client.unmarshalRequest(
		"POST",
		"replicasets",
		nil,
		bytes.NewReader(dataBytes),
		resp,
	)
	if status != nil || err != nil {
		return nil, status, err
	}
	return resp, nil, nil
}

// Patch patches the replicaset in `env` with `name`
func (s *ReplicaSetsService) Patch(env, name string, data *v1beta1.ReplicaSet) (*v1beta1.ReplicaSet, *unversioned.Status, error) {
	client, err := getClient(env)
	if err != nil {
		return nil, nil, invalidEnvError(env)
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, nil, err
	}
	resp := new(v1beta1.ReplicaSet)
	status, err := client.unmarshalRequest(
		"PATCH",
		"replicasets/"+name,
		nil,
		bytes.NewReader(dataBytes),
		resp,
	)
	if status != nil || err != nil {
		return nil, status, err
	}
	return resp, nil, nil
}

// Delete deletes the replicaset in `env` with `name`
func (s *ReplicaSetsService) Delete(env, name string) (*v1beta1.ReplicaSet, *unversioned.Status, error) {
	client, err := getClient(env)
	if err != nil {
		return nil, nil, invalidEnvError(env)
	}
	resp := new(v1beta1.ReplicaSet)
	status, err := client.unmarshalRequest("DELETE", "replicasets/"+name, nil, nil, resp)
	if status != nil || err != nil {
		return nil, status, err
	}
	return resp, nil, nil
}

// ReplicaSetEvent describes an event on a replica set
type ReplicaSetEvent struct {
	Type   WatchEventType          `json:"type"`
	Object *v1beta1.ReplicaSet     `json:"object"`
	List   *v1beta1.ReplicaSetList `json:"list"`
}

// Watch watches replicaSets in `env`
func (s *ReplicaSetsService) Watch(env string, query *url.Values) (watcher *Watcher, err error) {
	return watchObjectStream(env, "replicasets", query, func(eventType WatchEventType, body json.RawMessage) (interface{}, error) {
		if eventType == WatchEventInit {
			event := &ReplicaSetEvent{
				Type: eventType,
				List: new(v1beta1.ReplicaSetList),
			}
			return event, json.Unmarshal(body, event.List)
		}
		event := &ReplicaSetEvent{
			Type:   eventType,
			Object: new(v1beta1.ReplicaSet),
		}
		return event, json.Unmarshal(body, event.Object)
	})
}
