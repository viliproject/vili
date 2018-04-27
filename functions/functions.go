package functions

import (
	"context"
	"time"
)

var service Service

// List returns a list of functions for the given env
func List(ctx context.Context, env string) ([]Function, error) {
	return service.List(ctx, env)
}

// Get returns the given function for the given env
func Get(ctx context.Context, env, name string) (Function, error) {
	return service.Get(ctx, env, name)
}

// Deploy deploys the given function for the given env with the given spec
func Deploy(ctx context.Context, env, name string, spec *FunctionDeploySpec) error {
	return service.Deploy(ctx, env, name, spec)
}

// Rollback rolls the given function for the given env back to the given version
func Rollback(ctx context.Context, env, name, version string) error {
	return service.Rollback(ctx, env, name, version)
}

// Service is a functions service instance that manages the state of functions
// in the backing infrastructure
type Service interface {
	List(ctx context.Context, env string) ([]Function, error)
	Get(ctx context.Context, env, name string) (Function, error)
	Deploy(ctx context.Context, env, name string, spec *FunctionDeploySpec) error
	Rollback(ctx context.Context, env, name, version string) error
}

// Function is a function defined in the backing infrastructure
type Function interface {
	GetName() string
	GetEnv() string
	GetActiveVersion() FunctionVersion
	GetVersions() []FunctionVersion
}

// FunctionVersion is a function defined in the backing infrastructure
type FunctionVersion interface {
	GetVersion() string
	GetLastModified() time.Time
	GetEnv() string
	GetTag() string
	GetBranch() string
	GetDeployedBy() string
}

// NotFoundError is returned when a function is not found
type NotFoundError struct{}

func (NotFoundError) Error() string {
	return "Function not found"
}

// FunctionDeploySpec is used to deploy a function
type FunctionDeploySpec struct {
	Tag        string `json:"tag"`
	Branch     string `json:"branch"`
	DeployedBy string `json:"deployedBy"`
}
