package environments

import (
	"errors"
	"sort"
	"sync"

	"github.com/airware/vili/kube"
	"github.com/airware/vili/kube/v1"
)

var environments []Environment
var rwMutex sync.RWMutex

// Environment describes an environment backed by a kubernetes namespace
type Environment struct {
	Name      string `json:"name"`
	Protected bool   `json:"protected"`
	Prod      bool   `json:"prod"`
	Approval  bool   `json:"approval"`
}

// Init initializes the global environments list
func Init(envs []Environment) {
	rwMutex.Lock()
	environments = []Environment{}
	for _, env := range envs {
		environments = append(environments, env)
	}
	sort.Sort(byName(environments))
	rwMutex.Unlock()
}

// Environments returns a snapshot of all of the known environments
func Environments() (ret []Environment) {
	rwMutex.RLock()
	for _, env := range environments {
		ret = append(ret, env)
	}
	rwMutex.RUnlock()
	return
}

// Create creates a new environment with `name`
func Create(name string) error {
	_, status, err := kube.Namespaces.Create(&v1.Namespace{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
			Annotations: map[string]string{
				"airware.feature-environment": name,
			},
		},
	})
	if err != nil {
		return err
	}
	if status != nil {
		return errors.New(status.Message)
	}

	rwMutex.Lock()
	defer rwMutex.Unlock()
	environments = append(environments, Environment{
		Name: name,
	})
	sort.Sort(byName(environments))
	return nil
}

// Delete deletes the environment with `name`
func Delete(name string) error {
	rwMutex.Lock()
	defer rwMutex.Unlock()
	envIndex := -1
	for i, env := range environments {
		if env.Name == name {
			envIndex = i
			if env.Protected {
				return errors.New(name + " is a protected environment")
			}
		}
	}
	if envIndex < 0 {
		return errors.New(name + " not found")
	}

	status, err := kube.Namespaces.Delete(name)
	if err != nil {
		return err
	}
	if status != nil {
		return errors.New(status.Message)
	}

	environments = append(environments[:envIndex], environments[envIndex+1:]...)

	return nil
}

// RefreshEnvs refreshes the list of environments as detected from the kubernetes cluster
func RefreshEnvs() error {
	namespaceList, _, err := kube.Namespaces.List(nil)
	if err != nil {
		return err
	}

	newEnvs := []Environment{}
	rwMutex.Lock()
	defer rwMutex.Unlock()
	for _, env := range environments {
		if env.Protected {
			newEnvs = append(newEnvs, env)
		}
	}
	for _, namespace := range namespaceList.Items {
		if namespace.Name != "kube-system" && namespace.Name != "default" {
			newEnvs = append(newEnvs, Environment{
				Name: namespace.Name,
			})
		}
	}
	sort.Sort(byName(newEnvs))
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
