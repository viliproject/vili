package kube

import (
	"bytes"
	"io"
	"os/exec"
	"strings"
)

func kubectl(stdin io.Reader, args ...string) (string, error) {
	kubeArgs := []string{}
	if config.DefaultKubeConfigPath != "" {
		kubeArgs = append(kubeArgs, "--kubeconfig", config.DefaultKubeConfigPath)
	} else {
		kubeArgs = append(kubeArgs, "--server", defaultRestConfig.Host)
		if defaultRestConfig.BearerToken != "" {
			kubeArgs = append(kubeArgs, "--token", defaultRestConfig.BearerToken)
		}
		if defaultRestConfig.TLSClientConfig.CAFile != "" {
			kubeArgs = append(kubeArgs, "--certificate-authority", defaultRestConfig.TLSClientConfig.CAFile)
		}
	}

	kubeArgs = append(kubeArgs, args...)

	cmd := exec.Command("kubectl", kubeArgs...)
	cmd.Stdin = stdin

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// Create uses `kubectl create` to create the objects defined by `spec`
func Create(spec string) (map[string][]string, error) {
	out, err := kubectl(bytes.NewReader([]byte(spec)), "create", "-f", "-", "-o", "name")
	if err != nil {
		return nil, err
	}
	resources := make(map[string][]string)
	for _, resource := range strings.Fields(out) {
		parts := strings.SplitN(resource, "/", 2)
		resources[parts[0]] = append(resources[parts[0]], parts[1])
	}
	return resources, err
}

// Delete uses `kubectl delete` to delete the objects defined by `spec`
func Delete(spec string) (map[string][]string, error) {
	out, err := kubectl(bytes.NewReader([]byte(spec)), "delete", "-f", "-", "-o", "name")
	resources := make(map[string][]string)
	for _, resource := range strings.Fields(out) {
		parts := strings.SplitN(resource, "/", 2)
		resources[parts[0]] = append(resources[parts[0]], parts[1])
	}
	return resources, err
}
