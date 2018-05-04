package api

import (
	"sync"

	"github.com/airware/vili/environments"
	"github.com/airware/vili/errors"
	"github.com/airware/vili/middleware"
	"github.com/airware/vili/server"
	"github.com/labstack/echo"
)

// WaitGroup is the wait group to synchronize deployment rollouts
var WaitGroup sync.WaitGroup

// ExitingChan is a flag indicating that the server is exiting
var ExitingChan = make(chan struct{})

// AddHandlers adds api handlers to the server
func AddHandlers(s *server.Server) {
	envPrefix := "/api/v1/envs/:env/"
	// deployments
	s.Echo().GET(envPrefix+"deployments", envMiddleware(deploymentsGetHandler))
	s.Echo().GET(envPrefix+"deployments/:deployment/repository", envMiddleware(deploymentRepositoryGetHandler))
	s.Echo().GET(envPrefix+"deployments/:deployment/spec", envMiddleware(deploymentSpecGetHandler))
	s.Echo().GET(envPrefix+"deployments/:deployment/service", envMiddleware(deploymentServiceGetHandler))
	s.Echo().POST(envPrefix+"deployments/:deployment/service", envMiddleware(deploymentServiceCreateHandler))
	s.Echo().PUT(envPrefix+"deployments/:deployment/:action", envMiddleware(deploymentActionHandler))

	// rollouts
	s.Echo().POST(envPrefix+"deployments/:deployment/rollouts", envMiddleware(rolloutCreateHandler))

	// replica sets
	s.Echo().GET(envPrefix+"replicasets", envMiddleware(replicaSetsGetHandler))

	// jobs
	s.Echo().GET(envPrefix+"jobs", envMiddleware(jobsGetHandler))
	s.Echo().DELETE(envPrefix+"jobs/:job", envMiddleware(jobDeleteHandler))
	s.Echo().GET(envPrefix+"jobs/:job/repository", envMiddleware(jobRepositoryGetHandler))
	s.Echo().GET(envPrefix+"jobs/:job/spec", envMiddleware(jobSpecGetHandler))

	// runs
	s.Echo().POST(envPrefix+"jobs/:job/runs", envMiddleware(jobRunCreateHandler))
	// s.Echo().GET(envPrefix+"jobs/:job/runs", envMiddleware(jobRunsGetHandler))
	// s.Echo().POST(envPrefix+"jobs/:job/runs/:run/:action", envMiddleware(jobRunActionHandler))

	// functions
	s.Echo().GET(envPrefix+"functions", envMiddleware(functionsGetHandler))
	s.Echo().GET(envPrefix+"functions/:function/repository", envMiddleware(functionRepositoryGetHandler))
	s.Echo().GET(envPrefix+"functions/:function/spec", envMiddleware(functionSpecGetHandler))
	s.Echo().PUT(envPrefix+"functions/:function/:action", envMiddleware(functionActionHandler))

	// configmaps
	s.Echo().GET(envPrefix+"configmaps", envMiddleware(configmapsGetHandler))
	s.Echo().GET(envPrefix+"configmaps/:configmap/spec", envMiddleware(configmapSpecGetHandler))
	s.Echo().POST(envPrefix+"configmaps/:configmap", envMiddleware(configmapCreateHandler))
	s.Echo().DELETE(envPrefix+"configmaps/:configmap", envMiddleware(configmapDeleteHandler))
	s.Echo().PUT(envPrefix+"configmaps/:configmap/keys", envMiddleware(configmapSetKeysHandler))
	s.Echo().DELETE(envPrefix+"configmaps/:configmap/:key", envMiddleware(configmapDeleteKeyHandler))

	// pods
	s.Echo().GET(envPrefix+"pods", envMiddleware(podsHandler))
	s.Echo().GET(envPrefix+"pods/:pod/log", envMiddleware(podLogHandler))
	s.Echo().DELETE(envPrefix+"pods/:pod", envMiddleware(podDeleteHandler))

	// nodes
	s.Echo().GET(envPrefix+"nodes", envMiddleware(nodesGetHandler))
	s.Echo().PUT(envPrefix+"nodes/:node/:state", envMiddleware(nodeStateEditHandler))

	// releases
	s.Echo().GET(envPrefix+"releases", envMiddleware(releasesGetHandler))
	s.Echo().GET(envPrefix+"releases/spec", envMiddleware(releaseSpecGetHandler))
	s.Echo().POST(envPrefix+"releases", envMiddleware(releaseCreateHandler))
	s.Echo().DELETE(envPrefix+"releases/:release", envMiddleware(releaseDeleteHandler))
	s.Echo().PUT(envPrefix+"releases/:release/deploy", envMiddleware(releaseDeployHandler))

	// branches
	s.Echo().GET("/api/v1/branches", middleware.RequireUser(branchesGetHandler))

	// environments
	s.Echo().GET("/api/v1/environments", middleware.RequireUser(environmentsGetHandler))
	s.Echo().POST("/api/v1/environments", middleware.RequireUser(environmentCreateHandler))
	s.Echo().DELETE("/api/v1/environments/:env", middleware.RequireUser(environmentDeleteHandler))
	s.Echo().GET("/api/v1/environments/spec", middleware.RequireUser(environmentSpecHandler))

	// catchall not found handler
	s.Echo().GET("/api/**", middleware.RequireUser(notFoundHandler))
}

func envMiddleware(h echo.HandlerFunc) echo.HandlerFunc {
	return middleware.RequireUser(func(c echo.Context) error {
		if _, err := environments.Get(c.Param("env")); err != nil {
			return notFoundHandler(c)
		}
		return h(c)
	})
}

func notFoundHandler(c echo.Context) error {
	return server.ErrorResponse(c, errors.NotFound(""))
}
