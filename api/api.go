package api

import (
	"sync"

	"github.com/airware/vili/errors"
	"github.com/airware/vili/middleware"
	"github.com/airware/vili/server"
	"github.com/airware/vili/util"
	"github.com/labstack/echo"
)

// WaitGroup is the wait group to synchronize deployments and job runs
var WaitGroup sync.WaitGroup

// Exiting is a flag indicating that the server is exiting
var Exiting = false

// AddHandlers adds api handlers to the server
func AddHandlers(s *server.Server, envs *util.StringSet) {
	envPrefix := "/api/v1/envs/:env/"
	// apps
	s.Echo().Get(envPrefix+"apps", envMiddleware(envs, appsHandler))
	s.Echo().Get(envPrefix+"apps/:app", envMiddleware(envs, appHandler))
	s.Echo().Post(envPrefix+"apps/:app/service", envMiddleware(envs, appCreateServiceHandler))
	s.Echo().Put(envPrefix+"apps/:app/scale", envMiddleware(envs, appScaleHandler))

	// jobs
	s.Echo().Get(envPrefix+"jobs/:job", envMiddleware(envs, jobHandler))

	// nodes
	s.Echo().Get(envPrefix+"nodes", envMiddleware(envs, nodesHandler))
	s.Echo().Get(envPrefix+"nodes/:node", envMiddleware(envs, nodeHandler))
	s.Echo().Put(envPrefix+"nodes/:node/:state", envMiddleware(envs, nodeStateEditHandler))

	// pods
	s.Echo().Get(envPrefix+"pods", envMiddleware(envs, podsHandler))
	s.Echo().Get(envPrefix+"pods/:pod", envMiddleware(envs, podHandler))
	s.Echo().Delete(envPrefix+"pods/:pod", envMiddleware(envs, podDeleteHandler))

	// deployments
	s.Echo().Post(envPrefix+"apps/:app/deployments", envMiddleware(envs, deploymentCreateHandler))
	s.Echo().Put(envPrefix+"apps/:app/deployments/:deployment/rollout", envMiddleware(envs, deploymentRolloutEditHandler))
	s.Echo().Post(envPrefix+"apps/:app/deployments/:deployment/:action", envMiddleware(envs, deploymentActionHandler))

	// runs
	s.Echo().Post(envPrefix+"jobs/:job/runs", envMiddleware(envs, runCreateHandler))
	s.Echo().Post(envPrefix+"jobs/:job/runs/:run/:action", envMiddleware(envs, runActionHandler))

	// releases
	s.Echo().Post("/api/v1/releases/:app/:tag", middleware.RequireUser(releaseCreateHandler))
	s.Echo().Delete("/api/v1/releases/:app/:tag", middleware.RequireUser(releaseDeleteHandler))

	// environments
	s.Echo().Put("/api/v1/environments/:env", middleware.RequireUser(environmentCreateHandler))
	s.Echo().Delete("/api/v1/environments/:env", middleware.RequireUser(environmentDeleteHandler))

	// catchall not found handler
	s.Echo().Get("/api/*", middleware.RequireUser(notFoundHandler))
}

func envMiddleware(envs *util.StringSet, h echo.HandlerFunc) echo.HandlerFunc {
	return middleware.RequireUser(func(c *echo.Context) error {
		if !envs.Contains(c.Param("env")) {
			return notFoundHandler(c)
		}
		return h(c)
	})
}

func notFoundHandler(c *echo.Context) error {
	return server.ErrorResponse(c, errors.NotFoundError(""))
}
