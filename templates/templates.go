package templates

import (
	"bytes"
	"regexp"

	"github.com/airware/vili/kube/yaml"
)

var service Service

// Service is a template service that returns controller and pod templates for given
// environments
type Service interface {
	Deployments(env, branch string) ([]string, error)
	Deployment(env, branch, name string) (Template, error)
	Pods(env, branch string) ([]string, error)
	Pod(env, branch, name string) (Template, error)
	Environment(branch string) (Template, error)
	Variables(env, branch string) (map[string]string, error)
}

// Deployments returns a list of deployments for the given environment
func Deployments(env, branch string) ([]string, error) {
	return service.Deployments(env, branch)
}

// Deployment returns a deployment for the given environment
func Deployment(env, branch, name string) (Template, error) {
	return service.Deployment(env, branch, name)
}

// Pods returns a list of pods for the given environment
func Pods(env, branch string) ([]string, error) {
	return service.Pods(env, branch)
}

// Pod returns a list of pods for the given environment
func Pod(env, branch, name string) (Template, error) {
	return service.Pod(env, branch, name)
}

// Environment returns an environment template for the given branch
func Environment(branch string) (Template, error) {
	return service.Environment(branch)
}

// Variables returns a list of variabless for the given environment
func Variables(env, branch string) (map[string]string, error) {
	return service.Variables(env, branch)
}

// Template is a yaml string template of a controller of a pod
type Template string

var templateVariableRegexp = regexp.MustCompile("\\{[a-zA-Z0-9_]+\\}")

// Populate populates the template with variables and returns a new Template instance
func (t Template) Populate(variables map[string]string) (Template, bool) {
	invalid := false
	return Template(templateVariableRegexp.ReplaceAllFunc([]byte(t), func(match []byte) []byte {
		varname := string(match[1 : len(match)-1])
		val, ok := variables[varname]
		if !ok {
			invalid = true
		}
		return []byte(val)
	})), invalid
}

// Parse parses template into the given interface
func (t Template) Parse(into interface{}) error {
	decoder := yaml.NewYAMLToJSONDecoder(bytes.NewReader([]byte(t)))
	return decoder.Decode(into)
}
