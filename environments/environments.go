package environments

import (
	"errors"
	"os/exec"
	"sort"
	"sync"

	"github.com/airware/vili/config"
	"github.com/airware/vili/kube"
	"github.com/airware/vili/log"
	"github.com/airware/vili/templates"
	"github.com/airware/vili/util"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

var (
	// ExitingChan is a flag indicating that the server is exiting
	ExitingChan   = make(chan struct{})
	environments  map[string]*Environment
	namespaceEnvs map[string]string
	ignoredEnvs   *util.StringSet
	rwMutex       sync.RWMutex
)

// Environment describes an environment backed by a kubernetes namespace
type Environment struct {
	Name               string   `json:"name"`
	Branch             string   `json:"branch,omitempty"`
	RepositoryBranches []string `json:"repositoryBranches,omitempty"`
	AutodeployBranches []string `json:"autodeployBranches,omitempty"`
	Protected          bool     `json:"protected,omitempty"`
	DeployedToEnv      string   `json:"deployedToEnv,omitempty"`
	ApprovedFromEnv    string   `json:"approvedFromEnv,omitempty"`
	Jobs               []string `json:"jobs"`
	Deployments        []string `json:"deployments"`
	ConfigMaps         []string `json:"configmaps"`
}

func (e *Environment) fillBranches() {
	defaultBranch := "develop"
	if e.ApprovedFromEnv != "" || e.DeployedToEnv != "" {
		defaultBranch = "master"
	}
	if e.Branch == "" {
		e.Branch = defaultBranch
	}
	e.RepositoryBranches = config.GetStringSlice(config.EnvRepositoryBranches(e.Name))
	if !util.NewStringSet(e.RepositoryBranches).Contains(e.Branch) {
		e.RepositoryBranches = append(e.RepositoryBranches, e.Branch)
	}
	if !util.NewStringSet(e.RepositoryBranches).Contains(defaultBranch) {
		e.RepositoryBranches = append(e.RepositoryBranches, defaultBranch)
	}
	e.AutodeployBranches = []string{e.Branch}
	if defaultBranch == "master" && !util.NewStringSet(e.RepositoryBranches).Contains(defaultBranch) {
		e.AutodeployBranches = append(e.AutodeployBranches, defaultBranch)
	}
}

func (e *Environment) fillSpecs() {
	jobs, err := templates.Jobs(e.Name, e.Branch)
	if err != nil {
		log.Error(err)
		return
	}
	deployments, err := templates.Deployments(e.Name, e.Branch)
	if err != nil {
		log.Error(err)
		return
	}
	configMaps, err := templates.ConfigMaps(e.Name, e.Branch)
	if err != nil {
		log.Error(err)
		return
	}
	e.Jobs = jobs
	e.Deployments = deployments
	e.ConfigMaps = configMaps
}

// Init initializes the global environments list
func Init() {
	rwMutex.Lock()
	environments = make(map[string]*Environment)
	ignoredEnvs = util.NewStringSet(append(config.GetStringSlice(config.IgnoredEnvs), "kube-system", "default"))
	envKubeNamespaces := config.GetStringSliceMap(config.EnvKubernetesNamespaces)
	namespaceEnvs = map[string]string{}
	for env, namespace := range envKubeNamespaces {
		namespaceEnvs[namespace] = env
	}

	deployedToEnvs := config.GetStringSliceMap(config.ApprovalProdEnvs)
	approvedFromEnvs := map[string]string{}
	for k, v := range deployedToEnvs {
		approvedFromEnvs[v] = k
	}
	for _, envName := range config.GetStringSlice(config.Environments) {
		env := &Environment{
			Name:            envName,
			Protected:       true,
			DeployedToEnv:   deployedToEnvs[envName],
			ApprovedFromEnv: approvedFromEnvs[envName],
		}
		env.fillBranches()
		environments[env.Name] = env
	}
	rwMutex.Unlock()
}

// Environments returns a snapshot of all of the known environments
func Environments() (ret []*Environment) {
	rwMutex.RLock()
	for _, env := range environments {
		ret = append(ret, env)
	}
	rwMutex.RUnlock()
	sort.Sort(byProtectedAndName(ret))
	return
}

// Get returns the environment with `name`
func Get(name string) (*Environment, error) {
	rwMutex.RLock()
	env, ok := environments[name]
	rwMutex.RUnlock()
	if !ok {
		return nil, errors.New(name + " not found")
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

	env := &Environment{
		Name:   name,
		Branch: branch,
	}
	env.fillBranches()
	env.fillSpecs()
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

	return kube.GetClient("").Core().Namespaces().Delete(name, nil)
}

// WatchEnvs watches the namespaces on the kubernetes cluster and updates the list of environments
func WatchEnvs() {
	watcher, err := kube.GetClient("").Core().Namespaces().Watch(metav1.ListOptions{})
	if err != nil {
		log.WithError(err).Error("error watching namespaces")
		return
	}
	go func() {
		for {
			select {
			case event, ok := <-watcher.ResultChan():
				if !ok {
					log.Debug("namespace watcher disconnected, reconnecting...")
					watcher, err = kube.GetClient("").Core().Namespaces().Watch(metav1.ListOptions{})
					if err != nil {
						log.WithError(err).Error("error watching namespaces")
						return
					}
				} else {
					namespace := event.Object.(*apiv1.Namespace)
					switch event.Type {
					case watch.Added, watch.Modified:
						updateEnv(namespace)
					case watch.Deleted:
						rwMutex.Lock()
						delete(environments, namespace.Name)
						rwMutex.Unlock()
					}
				}
			case <-ExitingChan:
				watcher.Stop()
				return
			}
		}
	}()
	<-ExitingChan
	if watcher != nil {
		log.Info("stopping namespace watcher")
		watcher.Stop()
	}
}

func updateEnv(namespace *apiv1.Namespace) {
	envName := namespace.Name
	if namespaceEnvs[envName] != "" {
		envName = namespaceEnvs[envName]
	}
	if !ignoredEnvs.Contains(envName) {
		if namespace.Status.Phase == "Terminating" {
			rwMutex.Lock()
			delete(environments, envName)
			rwMutex.Unlock()
		} else {
			env, ok := environments[envName]
			if ok {
				env.Branch = namespace.Annotations["vili.environment-branch"]
			} else {
				env = &Environment{
					Name:   envName,
					Branch: namespace.Annotations["vili.environment-branch"],
				}
			}
			env.fillBranches()
			env.fillSpecs()
			rwMutex.Lock()
			environments[envName] = env
			rwMutex.Unlock()
		}
	}
}

type byProtectedAndName []*Environment

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
