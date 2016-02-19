package kube

import (
	"bytes"
	"encoding/json"
	"net/url"

	"github.com/airware/vili/kube/unversioned"
	"github.com/airware/vili/kube/v1"
)

// Pods is the default nodes service instance
var Nodes = &NodesService{}

// NodesService is the kubernetes service to interace with nodes
type NodesService struct {
}

// List fetches the list of nodes in `env`
func (s *NodesService) List(env string, query *url.Values) (*v1.NodeList, error) {
	envConfig := config.EnvConfigs[env]
	if envConfig == nil {
		return nil, invalidEnvError(env)
	}
	resp := &v1.NodeList{}
	path := "nodes"
	if query != nil {
		path += "?" + query.Encode()
	}
	_, err := envConfig.client.makeRequest("GET", path, nil, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// Get fetches the node in `env` with `name`
func (s *NodesService) Get(env, name string) (*v1.Node, *unversioned.Status, error) {
	envConfig := config.EnvConfigs[env]
	if envConfig == nil {
		return nil, nil, invalidEnvError(env)
	}
	resp := &v1.Node{}
	status, err := envConfig.client.makeRequest("GET", "nodes/"+name, nil, resp)
	if status != nil || err != nil {
		return nil, status, err
	}
	return resp, nil, nil
}

// Patch patches the node in `env` with `name`
func (s *NodesService) Patch(env, name string, data *v1.Node) (*v1.Node, error) {
	envConfig := config.EnvConfigs[env]
	if envConfig == nil {
		return nil, invalidEnvError(env)
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	resp := &v1.Node{}
	_, err = envConfig.client.makeRequest(
		"PATCH",
		"nodes/"+name,
		bytes.NewReader(dataBytes),
		resp,
	)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// PatchUnschedulable changes the unschedulable value of the node in `env` with `name`
// This is necessary because Go doesn't serialize false Unschedulable values in v1.NodeSpec
func (s *NodesService) PatchUnschedulable(env, name string, unschedulable bool) (*v1.Node, error) {
	envConfig := config.EnvConfigs[env]
	if envConfig == nil {
		return nil, invalidEnvError(env)
	}

	data := &Node{
		Spec: NodeSpec{
			Unschedulable: unschedulable,
		},
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	resp := &v1.Node{}
	_, err = envConfig.client.makeRequest(
		"PATCH",
		"nodes/"+name,
		bytes.NewReader(dataBytes),
		resp,
	)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// Node is a custom struct representing a kubernetes Node
type Node struct {
	Spec NodeSpec `json:"spec,omitempty"`
}

// NodeSpec is a custom struct representing a kubernetes NodeSpec
type NodeSpec struct {
	Unschedulable bool `json:"unschedulable"`
}
