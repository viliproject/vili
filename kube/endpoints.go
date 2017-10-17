package kube

import (
	batchv1 "k8s.io/client-go/kubernetes/typed/batch/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	extensionsv1beta1 "k8s.io/client-go/kubernetes/typed/extensions/v1beta1"
)

// Services returns the services endpoint for the client's namespace
func (k *Client) Services() corev1.ServiceInterface {
	return k.Core().Services(k.namespace)
}

// Nodes returns the nodes endpoint for the client's namespace
func (k *Client) Nodes() corev1.NodeInterface {
	return k.Core().Nodes()
}

// Pods returns the pods endpoint for the client's namespace
func (k *Client) Pods() corev1.PodInterface {
	return k.Core().Pods(k.namespace)
}

// ConfigMaps returns the configMaps endpoint for the client's namespace
func (k *Client) ConfigMaps() corev1.ConfigMapInterface {
	return k.Core().ConfigMaps(k.namespace)
}

// Deployments returns the deployments endpoint for the client's namespace
func (k *Client) Deployments() extensionsv1beta1.DeploymentInterface {
	return k.Extensions().Deployments(k.namespace)
}

// Jobs returns the jobs endpoint for the client's namespace
func (k *Client) Jobs() batchv1.JobInterface {
	return k.Batch().Jobs(k.namespace)
}

// ReplicaSets returns the replicaSets endpoint for the client's namespace
func (k *Client) ReplicaSets() extensionsv1beta1.ReplicaSetInterface {
	return k.Extensions().ReplicaSets(k.namespace)
}
