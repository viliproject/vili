package environments

import (
	"errors"
	"os/exec"
	"sort"
	"sync"

	"github.com/airware/vili/kube"
)

var environments map[string]Environment
var rwMutex sync.RWMutex

// Environment describes an environment backed by a kubernetes namespace
type Environment struct {
	Name      string `json:"name"`
	Branch    string `json:"branch,omitempty"`
	Protected bool   `json:"protected,omitempty"`
	Prod      bool   `json:"prod,omitempty"`
	Approval  bool   `json:"approval,omitempty"`
}

// Init initializes the global environments list
func Init(envs []Environment) {
	rwMutex.Lock()
	environments = make(map[string]Environment)
	for _, env := range envs {
		environments[env.Name] = env
	}
	rwMutex.Unlock()
}

// Environments returns a snapshot of all of the known environments
func Environments() (ret []Environment) {
	rwMutex.RLock()
	for _, env := range environments {
		ret = append(ret, env)
	}
	rwMutex.RUnlock()
	sort.Sort(byName(ret))
	return
}

// Get returns the environment with `name`
func Get(name string) (Environment, error) {
	rwMutex.RLock()
	defer rwMutex.RUnlock()
	env, ok := environments[name]
	if !ok {
		return Environment{}, errors.New(name + " not found")
	}
	return env, nil
}

// Create creates a new environment with `name`
func Create(name, branch, spec string) (map[string][]string, error) {
	resources, err := kube.Create(spec)
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, errors.New(string(exitErr.Stderr))
		}
		return nil, err
	}

	rwMutex.Lock()
	environments[name] = Environment{
		Name:   name,
		Branch: branch,
	}
	rwMutex.Unlock()
	return resources, nil
}

// Delete deletes the environment with `name`
func Delete(name string) error {
	rwMutex.Lock()
	defer rwMutex.Unlock()

	env, ok := environments[name]
	if !ok {
		return errors.New(name + " not found")
	} else if env.Protected {
		return errors.New(name + " is a protected environment")
	}

	status, err := kube.Namespaces.Delete(name)
	if err != nil {
		return err
	}
	if status != nil {
		return errors.New(status.Message)
	}

	delete(environments, name)
	return nil
}

// RefreshEnvs refreshes the list of environments as detected from the kubernetes cluster
func RefreshEnvs() error {
	namespaceList, _, err := kube.Namespaces.List(nil)
	if err != nil {
		return err
	}

	newEnvs := make(map[string]Environment)
	rwMutex.Lock()
	defer rwMutex.Unlock()
	for name, env := range environments {
		if env.Protected {
			newEnvs[name] = env
		}
	}
	for _, namespace := range namespaceList.Items {
		if namespace.Name != "kube-system" && namespace.Name != "default" && namespace.Status.Phase != "Terminating" {
			if env, ok := newEnvs[namespace.Name]; ok {
				env.Branch = namespace.Annotations["vili.environment-branch"]
				newEnvs[namespace.Name] = env
			} else {
				newEnvs[namespace.Name] = Environment{
					Name:   namespace.Name,
					Branch: namespace.Annotations["vili.environment-branch"],
				}
			}
		}
	}
	environments = newEnvs
	return nil
}

type byName []Environment

// Len implements the sort interface
func (e byName) Len() int {
	return len(e)
}

// Less implements the sort interface
func (e byName) Less(i, j int) bool {
	return e[i].Name < e[j].Name
}

// Swap implements the sort interface
func (e byName) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}
