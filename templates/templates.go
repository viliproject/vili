package templates

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/airware/vili/kube/yaml"
)

var service Service

// Service is a template service that returns controller and pod templates for given
// environments
type Service interface {
	Deployments(env string) ([]string, error)
	Deployment(env, name string) (Template, error)
	Pods(env string) ([]string, error)
	Pod(env, name string) (Template, error)
	Variables(env string) (map[string]string, error)
}

// Deployments returns a list of deployments for the given environment
func Deployments(env string) ([]string, error) {
	return service.Deployments(env)
}

// Deployment returns a deployment for the given environment
func Deployment(env, name string) (Template, error) {
	return service.Deployment(env, name)
}

// Pods returns a list of pods for the given environment
func Pods(env string) ([]string, error) {
	return service.Pods(env)
}

// Pod returns a list of pods for the given environment
func Pod(env, name string) (Template, error) {
	return service.Pod(env, name)
}

// Variables returns a list of variabless for the given environment
func Variables(env string) (map[string]string, error) {
	return service.Variables(env)
}

// Template is a yaml string template of a controller of a pod
type Template string

var templateVariableRegexp = regexp.MustCompile("\\{[a-zA-Z0-9_]+(:[a-zA-Z0-9_]*)?\\}")

// Populate populates the template with variables and returns a new Template instance
func (t Template) Populate(variables map[string]string) (Template, bool) {
	invalid := false
	return Template(templateVariableRegexp.ReplaceAllStringFunc(string(t), func(match string) string {
		var defaultVal *string
		varName := match[1 : len(match)-1]
		splitVar := strings.SplitN(varName, ":", 2)
		if len(splitVar) == 2 {
			varName = splitVar[0]
			defaultVal = &splitVar[1]
		}

		val, ok := variables[varName]
		if !ok {
			if defaultVal == nil {
				invalid = true
			} else {
				val = *defaultVal
			}
		}
		return val
	})), invalid
}

// ExtractVariables returns the variable names and default values from a template
func (t Template) ExtractVariables() map[string]string {
	variables := make(map[string]string)
	vars := templateVariableRegexp.FindAllString(string(t), -1)
	for _, v := range vars {
		varName := v[1 : len(v)-1]
		splitVar := strings.SplitN(varName, ":", 2)
		if len(splitVar) == 2 {
			variables[splitVar[0]] = splitVar[1]
		} else {
			variables[varName] = ""
		}
	}
	return variables
}

// Parse parses template into the given interface
func (t Template) Parse(into interface{}) error {
	decoder := yaml.NewYAMLToJSONDecoder(bytes.NewReader([]byte(t)))
	return decoder.Decode(into)
}
