package api

import (
	"sync"

	"github.com/airware/vili/environments"
	"github.com/airware/vili/errors"
	"github.com/airware/vili/middleware"
	"github.com/airware/vili/server"
	"gopkg.in/labstack/echo.v1"
)

// WaitGroup is the wait group to synchronize deployments and job runs
var WaitGroup sync.WaitGroup

// Exiting is a flag indicating that the server is exiting
var Exiting = false

// AddHandlers adds api handlers to the server
func AddHandlers(s *server.Server) {
	envPrefix := "/api/v1/envs/:env/"
	// apps
	s.Echo().Get(envPrefix+"apps", envMiddleware(appsHandler))
	s.Echo().Get(envPrefix+"apps/:app", envMiddleware(appHandler))
	s.Echo().Post(envPrefix+"apps/:app/service", envMiddleware(appCreateServiceHandler))
	s.Echo().Put(envPrefix+"apps/:app/scale", envMiddleware(appScaleHandler))

	// jobs
	s.Echo().Get(envPrefix+"jobs/:job", envMiddleware(jobHandler))

	// nodes
	s.Echo().Get(envPrefix+"nodes", envMiddleware(nodesHandler))
	s.Echo().Get(envPrefix+"nodes/:node", envMiddleware(nodeHandler))
	s.Echo().Put(envPrefix+"nodes/:node/:state", envMiddleware(nodeStateEditHandler))

	// pods
	s.Echo().Get(envPrefix+"pods", envMiddleware(podsHandler))
	s.Echo().Get(envPrefix+"pods/:pod", envMiddleware(podHandler))
	s.Echo().Delete(envPrefix+"pods/:pod", envMiddleware(podDeleteHandler))

	// deployments
	s.Echo().Post(envPrefix+"apps/:app/deployments", envMiddleware(deploymentCreateHandler))
	s.Echo().Put(envPrefix+"apps/:app/deployments/:deployment/rollout", envMiddleware(deploymentRolloutEditHandler))
	s.Echo().Post(envPrefix+"apps/:app/deployments/:deployment/:action", envMiddleware(deploymentActionHandler))

	// runs
	s.Echo().Post(envPrefix+"jobs/:job/runs", envMiddleware(runCreateHandler))
	s.Echo().Post(envPrefix+"jobs/:job/runs/:run/:action", envMiddleware(runActionHandler))

	// releases
	s.Echo().Post("/api/v1/releases/:app/:tag", middleware.RequireUser(releaseCreateHandler))
	s.Echo().Delete("/api/v1/releases/:app/:tag", middleware.RequireUser(releaseDeleteHandler))

	// environments
	s.Echo().Get("/api/v1/envSpec", middleware.RequireUser(environmentSpecHandler))
	s.Echo().Post("/api/v1/environments", middleware.RequireUser(environmentCreateHandler))
	s.Echo().Delete("/api/v1/environments/:env", middleware.RequireUser(environmentDeleteHandler))

	// catchall not found handler
	s.Echo().Get("/api/*", middleware.RequireUser(notFoundHandler))
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
	return server.ErrorResponse(c, errors.NotFoundError(""))
}
