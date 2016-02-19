package kube

import (
	"bytes"
	"encoding/json"
	"net/url"

	"github.com/airware/vili/kube/unversioned"
	"github.com/airware/vili/kube/v1"
)

// Controllers is the default controllers service instance
var Controllers = &ControllersService{}

// ControllersService is the kubernetes service to interace with controllers
type ControllersService struct {
}

// List fetches the list of controllers in `env`
func (s *ControllersService) List(env string, query *url.Values) (*v1.ReplicationControllerList, *unversioned.Status, error) {
	envConfig := config.EnvConfigs[env]
	if envConfig == nil {
		return nil, nil, invalidEnvError(env)
	}
	resp := &v1.ReplicationControllerList{}
	path := "replicationcontrollers"
	if query != nil {
		path += "?" + query.Encode()
	}
	status, err := envConfig.client.makeRequest("GET", path, nil, resp)
	if status != nil || err != nil {
		return nil, status, err
	}
	return resp, nil, nil
}

// Get fetches the controller in `env` with `name`
func (s *ControllersService) Get(env, name string) (*v1.ReplicationController, *unversioned.Status, error) {
	envConfig := config.EnvConfigs[env]
	if envConfig == nil {
		return nil, nil, invalidEnvError(env)
	}
	resp := &v1.ReplicationController{}
	status, err := envConfig.client.makeRequest("GET", "replicationcontrollers/"+name, nil, resp)
	if status != nil || err != nil {
		return nil, status, err
	}
	return resp, nil, nil
}

// Create creates a controller in `env`
func (s *ControllersService) Create(env string, data *v1.ReplicationController) (*v1.ReplicationController, *unversioned.Status, error) {
	envConfig := config.EnvConfigs[env]
	if envConfig == nil {
		return nil, nil, invalidEnvError(env)
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, nil, err
	}
	resp := &v1.ReplicationController{}
	status, err := envConfig.client.makeRequest(
		"POST",
		"replicationcontrollers",
		bytes.NewReader(dataBytes),
		resp,
	)
	if status != nil || err != nil {
		return nil, status, err
	}
	return resp, nil, nil
}

// Patch patches the controller in `env` with `name`
func (s *ControllersService) Patch(env, name string, data *v1.ReplicationController) (*v1.ReplicationController, *unversioned.Status, error) {
	envConfig := config.EnvConfigs[env]
	if envConfig == nil {
		return nil, nil, invalidEnvError(env)
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, nil, err
	}
	resp := &v1.ReplicationController{}
	status, err := envConfig.client.makeRequest(
		"PATCH",
		"replicationcontrollers/"+name,
		bytes.NewReader(dataBytes),
		resp,
	)
	if status != nil || err != nil {
		return nil, status, err
	}
	return resp, nil, nil
}

// Delete deletes the controller in `env` with `name`
func (s *ControllersService) Delete(env, name string) (*v1.ReplicationController, *unversioned.Status, error) {
	envConfig := config.EnvConfigs[env]
	if envConfig == nil {
		return nil, nil, invalidEnvError(env)
	}
	resp := &v1.ReplicationController{}
	status, err := envConfig.client.makeRequest("DELETE", "replicationcontrollers/"+name, nil, resp)
	if status != nil || err != nil {
		return nil, status, err
	}
	return resp, nil, nil
}
