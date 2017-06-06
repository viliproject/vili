package kube

import (
	"bytes"
	"encoding/json"

	"github.com/airware/vili/kube/unversioned"
	"github.com/airware/vili/kube/v1"
)

// Services is the default services service instance
var Services = new(ServicesService)

// ServicesService is the kubernetes service to interace with services
type ServicesService struct {
}

// List fetches the list of services in `env`
func (s *ServicesService) List(env string) (*v1.ServiceList, error) {
	client, err := getClient(env)
	if err != nil {
		return nil, invalidEnvError(env)
	}
	resp := new(v1.ServiceList)
	_, err = client.unmarshalRequest("GET", "replicationservices", nil, nil, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// Get fetches the service in `env` with `name`
func (s *ServicesService) Get(env, name string) (*v1.Service, *unversioned.Status, error) {
	client, err := getClient(env)
	if err != nil {
		return nil, nil, invalidEnvError(env)
	}
	resp := new(v1.Service)
	status, err := client.unmarshalRequest("GET", "services/"+name, nil, nil, resp)
	if status != nil || err != nil {
		return nil, status, err
	}
	return resp, nil, nil
}

// Create fetches the service in `env` with `name`
func (s *ServicesService) Create(env, name string, data *v1.Service) (*v1.Service, error) {
	client, err := getClient(env)
	if err != nil {
		return nil, invalidEnvError(env)
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	resp := new(v1.Service)
	_, err = client.unmarshalRequest(
		"POST",
		"services",
		nil,
		bytes.NewReader(dataBytes),
		resp,
	)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
