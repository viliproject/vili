package kube

import (
	"bytes"
	"encoding/json"
	"net/url"

	"github.com/airware/vili/kube/unversioned"
	"github.com/airware/vili/kube/v1"
)

// Namespaces is the default namespaces service instance
var Namespaces = &NamespacesService{}

// NamespacesService is the kubernetes service to interace with namespaces
type NamespacesService struct {
}

// List fetches the list of namespaces
func (s *NamespacesService) List(query *url.Values) (*v1.NamespaceList, *unversioned.Status, error) {
	client, err := getDefaultClient()
	if err != nil {
		return nil, nil, err
	}
	resp := &v1.NamespaceList{}
	path := "namespaces"
	if query != nil {
		path += "?" + query.Encode()
	}
	status, err := client.makeRequest("GET", path, nil, resp)
	if status != nil || err != nil {
		return nil, status, err
	}
	return resp, nil, nil
}

// Get fetches the namespace with `name`
func (s *NamespacesService) Get(name string) (*v1.Namespace, *unversioned.Status, error) {
	client, err := getDefaultClient()
	if err != nil {
		return nil, nil, err
	}
	resp := &v1.Namespace{}
	status, err := client.makeRequest("GET", "namespaces/"+name, nil, resp)
	if status != nil || err != nil {
		return nil, status, err
	}
	return resp, nil, nil
}

// Create creates a namespace
func (s *NamespacesService) Create(data *v1.Namespace) (*v1.Namespace, *unversioned.Status, error) {
	client, err := getDefaultClient()
	if err != nil {
		return nil, nil, err
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, nil, err
	}
	resp := &v1.Namespace{}
	status, err := client.makeRequest(
		"POST",
		"namespaces",
		bytes.NewReader(dataBytes),
		resp,
	)
	if status != nil || err != nil {
		return nil, status, err
	}
	return resp, nil, nil
}

// Delete deletes the namespace with `name`
func (s *NamespacesService) Delete(name string) (*unversioned.Status, error) {
	client, err := getDefaultClient()
	if err != nil {
		return nil, err
	}
	status, err := client.makeRequest("DELETE", "namespaces/"+name, nil, nil)
	if status != nil || err != nil {
		return status, err
	}
	return nil, nil
}
