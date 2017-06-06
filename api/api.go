package api

import (
	"sync"

	"github.com/airware/vili/environments"
	"github.com/airware/vili/errors"
	"github.com/airware/vili/middleware"
	"github.com/airware/vili/server"
	echo "gopkg.in/labstack/echo.v1"
)

// WaitGroup is the wait group to synchronize deployment rollouts
var WaitGroup sync.WaitGroup

// ExitingChan is a flag indicating that the server is exiting
var ExitingChan = make(chan struct{})

// AddHandlers adds api handlers to the server
func AddHandlers(s *server.Server) {
	envPrefix := "/api/v1/envs/:env/"
	// deployments
	s.Echo().Get(envPrefix+"deployments", envMiddleware(deploymentsGetHandler))
	s.Echo().Get(envPrefix+"deployments/:deployment/repository", envMiddleware(deploymentRepositoryGetHandler))
	s.Echo().Get(envPrefix+"deployments/:deployment/spec", envMiddleware(deploymentSpecGetHandler))
	s.Echo().Get(envPrefix+"deployments/:deployment/service", envMiddleware(deploymentServiceGetHandler))
	s.Echo().Post(envPrefix+"deployments/:deployment/service", envMiddleware(deploymentServiceCreateHandler))
	s.Echo().Put(envPrefix+"deployments/:deployment/:action", envMiddleware(deploymentActionHandler))

	// rollouts
	s.Echo().Post(envPrefix+"deployments/:deployment/rollouts", envMiddleware(rolloutCreateHandler))

	// replica sets
	s.Echo().Get(envPrefix+"replicasets", envMiddleware(replicaSetsGetHandler))

	// jobs
	s.Echo().Get(envPrefix+"jobs", envMiddleware(jobsGetHandler))
	s.Echo().Delete(envPrefix+"jobs/:job", envMiddleware(jobDeleteHandler))
	s.Echo().Get(envPrefix+"jobs/:job/repository", envMiddleware(jobRepositoryGetHandler))
	s.Echo().Get(envPrefix+"jobs/:job/spec", envMiddleware(jobSpecGetHandler))

	// runs
	s.Echo().Post(envPrefix+"jobs/:job/runs", envMiddleware(jobRunCreateHandler))
	// s.Echo().Get(envPrefix+"jobs/:job/runs", envMiddleware(jobRunsGetHandler))
	// s.Echo().Post(envPrefix+"jobs/:job/runs/:run/:action", envMiddleware(jobRunActionHandler))

	// configmaps
	s.Echo().Get(envPrefix+"configmaps", envMiddleware(configmapsGetHandler))
	s.Echo().Get(envPrefix+"configmaps/:configmap/spec", envMiddleware(configmapSpecGetHandler))
	s.Echo().Post(envPrefix+"configmaps/:configmap", envMiddleware(configmapCreateHandler))
	s.Echo().Delete(envPrefix+"configmaps/:configmap", envMiddleware(configmapDeleteHandler))
	s.Echo().Put(envPrefix+"configmaps/:configmap/keys", envMiddleware(configmapSetKeysHandler))
	s.Echo().Delete(envPrefix+"configmaps/:configmap/:key", envMiddleware(configmapDeleteKeyHandler))

	// pods
	s.Echo().Get(envPrefix+"pods", envMiddleware(podsHandler))
	s.Echo().Get(envPrefix+"pods/:pod/log", envMiddleware(podLogHandler))
	s.Echo().Delete(envPrefix+"pods/:pod", envMiddleware(podDeleteHandler))

	// nodes
	s.Echo().Get(envPrefix+"nodes", envMiddleware(nodesGetHandler))
	s.Echo().Put(envPrefix+"nodes/:node/:state", envMiddleware(nodeStateEditHandler))

	// releases
	s.Echo().Get(envPrefix+"releases", envMiddleware(releasesGetHandler))
	s.Echo().Get(envPrefix+"releases/spec", envMiddleware(releaseSpecGetHandler))
	s.Echo().Post(envPrefix+"releases", envMiddleware(releaseCreateHandler))
	s.Echo().Delete(envPrefix+"releases/:release", envMiddleware(releaseDeleteHandler))
	s.Echo().Put(envPrefix+"releases/:release/deploy", envMiddleware(releaseDeployHandler))

	// branches
	s.Echo().Get("/api/v1/branches", middleware.RequireUser(branchesGetHandler))

	// environments
	s.Echo().Post("/api/v1/environments", middleware.RequireUser(environmentCreateHandler))
	s.Echo().Delete("/api/v1/environments/:env", middleware.RequireUser(environmentDeleteHandler))
	s.Echo().Get("/api/v1/environments/spec", middleware.RequireUser(environmentSpecHandler))

	// catchall not found handler
	s.Echo().Get("/api/**", middleware.RequireUser(notFoundHandler))
}

func envMiddleware(h echo.HandlerFunc) echo.HandlerFunc {
	return middleware.RequireUser(func(c *echo.Context) error {
		if _, err := environments.Get(c.Param("env")); err != nil {
			return notFoundHandler(c)
		}
		return h(c)
	})
}

func notFoundHandler(c *echo.Context) error {
	return server.ErrorResponse(c, errors.NotFound(""))
}
