package kube

import (
	"bytes"
	"encoding/json"

	"github.com/airware/vili/kube/unversioned"
	"github.com/airware/vili/kube/v1"
)

// Services is the default services service instance
var Services = &ServicesService{}

// ServicesService is the kubernetes service to interace with services
type ServicesService struct {
}

// List fetches the list of services in `env`
func (s *ServicesService) List(env string) (*v1.ServiceList, error) {
	envConfig := config.EnvConfigs[env]
	if envConfig == nil {
		return nil, invalidEnvError(env)
	}
	resp := &v1.ServiceList{}
	_, err := envConfig.client.makeRequest("GET", "replicationservices", nil, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// Get fetches the service in `env` with `name`
func (s *ServicesService) Get(env, name string) (*v1.Service, *unversioned.Status, error) {
	envConfig := config.EnvConfigs[env]
	if envConfig == nil {
		return nil, nil, invalidEnvError(env)
	}
	resp := &v1.Service{}
	status, err := envConfig.client.makeRequest("GET", "services/"+name, nil, resp)
	if status != nil || err != nil {
		return nil, status, err
	}
	return resp, nil, nil
}

// Create fetches the service in `env` with `name`
func (s *ServicesService) Create(env, name string, data *v1.Service) (*v1.Service, error) {
	envConfig := config.EnvConfigs[env]
	if envConfig == nil {
		return nil, invalidEnvError(env)
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	resp := &v1.Service{}
	_, err = envConfig.client.makeRequest(
		"POST",
		"services",
		bytes.NewReader(dataBytes),
		resp,
	)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
