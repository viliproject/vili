package kube

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/airware/vili/kube/extensions/v1beta1"
	"github.com/airware/vili/kube/unversioned"
	"github.com/airware/vili/kube/v1"
)

// Pods is the default pods service instance
var Pods = &PodsService{}

// PodsService is the kubernetes service to interace with pods
type PodsService struct {
}

// List fetches the list of pods in `env`
func (s *PodsService) List(env string, query *url.Values) (*v1.PodList, *unversioned.Status, error) {
	client, err := getClient(env)
	if err != nil {
		return nil, nil, invalidEnvError(env)
	}
	resp := &v1.PodList{}
	path := "pods"
	if query != nil {
		path += "?" + query.Encode()
	}
	status, err := client.makeRequest("GET", path, nil, resp)
	if status != nil || err != nil {
		return nil, status, err
	}
	return resp, nil, nil
}

// ListForController fetches the list of pods in `env` for the given controller
func (s *PodsService) ListForController(env string, controller *v1.ReplicationController) (*v1.PodList, *unversioned.Status, error) {
	client, err := getClient(env)
	if err != nil {
		return nil, nil, invalidEnvError(env)
	}

	var selector []string
	for k, v := range controller.Spec.Selector {
		selector = append(selector, fmt.Sprintf("%s=%s", k, v))
	}
	resp := &v1.PodList{}
	path := fmt.Sprintf("pods?labelSelector=%s", strings.Join(selector, ","))
	status, err := client.makeRequest("GET", path, nil, resp)
	if status != nil || err != nil {
		return nil, status, err
	}
	return resp, nil, nil
}

// ListForReplicaSet fetches the list of pods in `env` for the given replicaset
func (s *PodsService) ListForReplicaSet(env string, replicaSet *v1beta1.ReplicaSet) (*v1.PodList, *unversioned.Status, error) {
	client, err := getClient(env)
	if err != nil {
		return nil, nil, invalidEnvError(env)
	}

	var selector []string
	for k, v := range replicaSet.Spec.Selector.MatchLabels {
		selector = append(selector, fmt.Sprintf("%s=%s", k, v))
	}
	resp := &v1.PodList{}
	path := fmt.Sprintf("pods?labelSelector=%s", strings.Join(selector, ","))
	status, err := client.makeRequest("GET", path, nil, resp)
	if status != nil || err != nil {
		return nil, status, err
	}
	return resp, nil, nil
}

// ListForDeployment fetches the list of pods in `env` for the given deployment
func (s *PodsService) ListForDeployment(env string, deployment *v1beta1.Deployment) (*v1.PodList, *unversioned.Status, error) {
	client, err := getClient(env)
	if err != nil {
		return nil, nil, invalidEnvError(env)
	}

	var selector []string
	for k, v := range deployment.Spec.Selector.MatchLabels {
		selector = append(selector, fmt.Sprintf("%s=%s", k, v))
	}
	resp := &v1.PodList{}
	path := fmt.Sprintf("pods?labelSelector=%s", strings.Join(selector, ","))
	status, err := client.makeRequest("GET", path, nil, resp)
	if status != nil || err != nil {
		return nil, status, err
	}
	return resp, nil, nil
}

// Get fetches the pod in `env` with `name`
func (s *PodsService) Get(env, name string) (*v1.Pod, *unversioned.Status, error) {
	client, err := getClient(env)
	if err != nil {
		return nil, nil, invalidEnvError(env)
	}
	resp := &v1.Pod{}
	status, err := client.makeRequest("GET", "pods/"+name, nil, resp)
	if status != nil || err != nil {
		return nil, status, err
	}
	return resp, nil, nil
}

// GetLog fetches the pod log in `env` with `name`
func (s *PodsService) GetLog(env, name string) (string, *unversioned.Status, error) {
	client, err := getClient(env)
	if err != nil {
		return "", nil, invalidEnvError(env)
	}
	body, status, err := client.makeRequestRaw("GET", "pods/"+name+"/log", nil)
	if status != nil || err != nil {
		return "", status, err
	}
	return string(body), nil, nil
}

// Create creates a pod in `env`
func (s *PodsService) Create(env string, data *v1.Pod) (*v1.Pod, *unversioned.Status, error) {
	client, err := getClient(env)
	if err != nil {
		return nil, nil, invalidEnvError(env)
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, nil, err
	}
	resp := &v1.Pod{}
	status, err := client.makeRequest(
		"POST",
		"pods",
		bytes.NewReader(dataBytes),
		resp,
	)
	if status != nil || err != nil {
		return nil, status, err
	}
	return resp, nil, nil
}

// Delete deletes the pod in `env` with `name`
func (s *PodsService) Delete(env, name string) (*v1.Pod, *unversioned.Status, error) {
	client, err := getClient(env)
	if err != nil {
		return nil, nil, invalidEnvError(env)
	}
	resp := &v1.Pod{}
	status, err := client.makeRequest("DELETE", "pods/"+name, nil, resp)
	if status != nil || err != nil {
		return nil, status, err
	}
	return resp, nil, nil
}

// DeleteForController deletes the pods in `env` for the given controller
func (s *PodsService) DeleteForController(env string, controller *v1.ReplicationController) (*v1.PodList, *unversioned.Status, error) {
	client, err := getClient(env)
	if err != nil {
		return nil, nil, invalidEnvError(env)
	}

	podList, status, err := s.ListForController(env, controller)
	if status != nil || err != nil {
		return nil, status, err
	}
	for _, pod := range podList.Items {
		resp := &v1.Pod{}
		status, err := client.makeRequest("DELETE", "pods/"+pod.ObjectMeta.Name, nil, resp)
		if status != nil || err != nil {
			return nil, status, err
		}
	}
	return podList, nil, nil
}
