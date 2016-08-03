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
var ReplicaSets = &ReplicaSetsService{}

// ReplicaSetsService is the kubernetes service to interace with replicasets
type ReplicaSetsService struct {
}

// List fetches the list of replicasets in `env`
func (s *ReplicaSetsService) List(env string, query *url.Values) (*v1beta1.ReplicaSetList, *unversioned.Status, error) {
	client, err := getClient(env)
	if err != nil {
		return nil, nil, invalidEnvError(env)
	}
	resp := &v1beta1.ReplicaSetList{}
	path := "replicasets"
	if query != nil {
		path += "?" + query.Encode()
	}
	status, err := client.makeRequest("GET", path, nil, resp)
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
	query := &url.Values{}
	query.Add("labelSelector", strings.Join(selector, ","))
	return s.List(env, query)
}

// Get fetches the replicaset in `env` with `name`
func (s *ReplicaSetsService) Get(env, name string) (*v1beta1.ReplicaSet, *unversioned.Status, error) {
	client, err := getClient(env)
	if err != nil {
		return nil, nil, invalidEnvError(env)
	}
	resp := &v1beta1.ReplicaSet{}
	status, err := client.makeRequest("GET", "replicasets/"+name, nil, resp)
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
	resp := &v1beta1.ReplicaSet{}
	status, err := client.makeRequest(
		"POST",
		"replicasets",
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
	resp := &v1beta1.ReplicaSet{}
	status, err := client.makeRequest(
		"PATCH",
		"replicasets/"+name,
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
	resp := &v1beta1.ReplicaSet{}
	status, err := client.makeRequest("DELETE", "replicasets/"+name, nil, resp)
	if status != nil || err != nil {
		return nil, status, err
	}
	return resp, nil, nil
}
