package environments

import (
	"errors"
	"os/exec"
	"sort"
	"sync"

	"github.com/airware/vili/kube"
	"github.com/airware/vili/log"
	"github.com/airware/vili/templates"
)

var environments map[string]Environment
var rwMutex sync.RWMutex

// Environment describes an environment backed by a kubernetes namespace
type Environment struct {
	Name      string   `json:"name"`
	Branch    string   `json:"branch,omitempty"`
	Protected bool     `json:"protected,omitempty"`
	Prod      bool     `json:"prod,omitempty"`
	Approval  bool     `json:"approval,omitempty"`
	Apps      []string `json:"apps"`
	Jobs      []string `json:"jobs"`
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
	sort.Sort(byProtectedAndName(ret))
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

	env := Environment{
		Name:   name,
		Branch: branch,
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		deployments, err := templates.Deployments(name, branch)
		if err != nil {
			log.Error(err)
		}
		env.Apps = deployments
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		pods, err := templates.Pods(name, branch)
		if err != nil {
			log.Error(err)
		}
		env.Jobs = pods
	}()

	wg.Wait()

	rwMutex.Lock()
	environments[name] = env
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

	var wg sync.WaitGroup
	var mapLock sync.Mutex
	for _, namespace := range namespaceList.Items {
		if namespace.Name != "kube-system" && namespace.Name != "default" && namespace.Status.Phase != "Terminating" {
			env, ok := newEnvs[namespace.Name]
			if ok {
				env.Branch = namespace.Annotations["vili.environment-branch"]
			} else {
				env = Environment{
					Name:   namespace.Name,
					Branch: namespace.Annotations["vili.environment-branch"],
				}
			}
			newEnvs[namespace.Name] = env
			wg.Add(1)
			go func(name, branch string) {
				defer wg.Done()
				deployments, err := templates.Deployments(name, branch)
				if err != nil {
					log.Error(err)
				}
				mapLock.Lock()
				env = newEnvs[name]
				env.Apps = deployments
				newEnvs[name] = env
				mapLock.Unlock()
			}(namespace.Name, env.Branch)
			wg.Add(1)
			go func(name, branch string) {
				defer wg.Done()
				pods, err := templates.Pods(name, branch)
				if err != nil {
					log.Error(err)
				}
				mapLock.Lock()
				env = newEnvs[name]
				env.Jobs = pods
				newEnvs[name] = env
				mapLock.Unlock()
			}(namespace.Name, env.Branch)
		}
	}
	wg.Wait()
	environments = newEnvs
	return nil
}

type byProtectedAndName []Environment

// Len implements the sort interface
func (e byProtectedAndName) Len() int {
	return len(e)
}

// Less implements the sort interface
func (e byProtectedAndName) Less(i, j int) bool {
	if e[i].Protected != e[j].Protected {
		return e[i].Protected
	}
	return e[i].Name < e[j].Name
}

// Swap implements the sort interface
func (e byProtectedAndName) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}
