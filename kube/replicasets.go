package kube

import (
	"bytes"
	"encoding/json"
	"net/url"

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
	envConfig := config.EnvConfigs[env]
	if envConfig == nil {
		return nil, nil, invalidEnvError(env)
	}
	resp := &v1beta1.ReplicaSetList{}
	path := "replicasets"
	if query != nil {
		path += "?" + query.Encode()
	}
	status, err := envConfig.client.makeRequest("GET", path, nil, resp)
	if status != nil || err != nil {
		return nil, status, err
	}
	return resp, nil, nil
}

// Get fetches the replicaset in `env` with `name`
func (s *ReplicaSetsService) Get(env, name string) (*v1beta1.ReplicaSet, *unversioned.Status, error) {
	envConfig := config.EnvConfigs[env]
	if envConfig == nil {
		return nil, nil, invalidEnvError(env)
	}
	resp := &v1beta1.ReplicaSet{}
	status, err := envConfig.client.makeRequest("GET", "replicasets/"+name, nil, resp)
	if status != nil || err != nil {
		return nil, status, err
	}
	return resp, nil, nil
}

// Create creates a replicaset in `env`
func (s *ReplicaSetsService) Create(env string, data *v1beta1.ReplicaSet) (*v1beta1.ReplicaSet, *unversioned.Status, error) {
	envConfig := config.EnvConfigs[env]
	if envConfig == nil {
		return nil, nil, invalidEnvError(env)
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, nil, err
	}
	resp := &v1beta1.ReplicaSet{}
	status, err := envConfig.client.makeRequest(
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
	envConfig := config.EnvConfigs[env]
	if envConfig == nil {
		return nil, nil, invalidEnvError(env)
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, nil, err
	}
	resp := &v1beta1.ReplicaSet{}
	status, err := envConfig.client.makeRequest(
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
	envConfig := config.EnvConfigs[env]
	if envConfig == nil {
		return nil, nil, invalidEnvError(env)
	}
	resp := &v1beta1.ReplicaSet{}
	status, err := envConfig.client.makeRequest("DELETE", "replicasets/"+name, nil, resp)
	if status != nil || err != nil {
		return nil, status, err
	}
	return resp, nil, nil
}
