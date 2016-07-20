package kube

import (
	"bytes"
	"encoding/json"
	"net/url"

	"github.com/airware/vili/kube/extensions/v1beta1"
	"github.com/airware/vili/kube/unversioned"
)

// Deployments is the default deployments service instance
var Deployments = &DeploymentsService{}

// DeploymentsService is the kubernetes service to interace with deployments
type DeploymentsService struct {
}

// List fetches the list of deployments in `env`
func (s *DeploymentsService) List(env string, query *url.Values) (*v1beta1.DeploymentList, *unversioned.Status, error) {
	client, err := getClient(env)
	if err != nil {
		return nil, nil, invalidEnvError(env)
	}
	resp := &v1beta1.DeploymentList{}
	path := "deployments"
	if query != nil {
		path += "?" + query.Encode()
	}
	status, err := client.makeRequest("GET", path, nil, resp)
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
	resp := &v1beta1.Deployment{}
	status, err := client.makeRequest("GET", "deployments/"+name, nil, resp)
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
	resp := &v1beta1.Deployment{}
	status, err := client.makeRequest(
		"POST",
		"deployments",
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
	resp := &v1beta1.Deployment{}
	status, err := client.makeRequest(
		"PUT",
		"deployments/"+name,
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
	resp := &v1beta1.Scale{}
	status, err := client.makeRequest(
		"PATCH",
		"deployments/"+name+"/scale",
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
	resp := &v1beta1.DeploymentRollback{}
	status, err := client.makeRequest(
		"POST",
		"deployments/"+name+"/rollback",
		bytes.NewReader(dataBytes),
		resp,
	)
	if status != nil || err != nil {
		return nil, status, err
	}
	return resp, nil, nil
}

// Delete deletes the deployment in `env` with `name`
func (s *DeploymentsService) Delete(env, name string) (*v1beta1.Deployment, *unversioned.Status, error) {
	client, err := getClient(env)
	if err != nil {
		return nil, nil, invalidEnvError(env)
	}
	resp := &v1beta1.Deployment{}
	status, err := client.makeRequest("DELETE", "deployments/"+name, nil, resp)
	if status != nil || err != nil {
		return nil, status, err
	}
	return resp, nil, nil
}
