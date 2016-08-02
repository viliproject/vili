// Package server defines the server's interface
package server

import (
	"expvar"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/airware/vili/errors"
	"github.com/airware/vili/log"
	"github.com/airware/vili/middleware"
	"github.com/tylerb/graceful"
	"gopkg.in/labstack/echo.v1"
	mw "gopkg.in/labstack/echo.v1/middleware"
)

// Server is an instance of the server
type Server struct {
	e *echo.Echo
	c *Config
	g *graceful.Server
	t *httptest.Server
}

// Config is the configuration for the server
type Config struct {
	Name         string
	Addr         string
	Timeout      time.Duration
	HealthCheck  func() error
	ShutdownFunc func()
	Middleware   []echo.MiddlewareFunc
}

// New returns a configured Server struct
func New(config *Config) *Server {
	e := echo.New()

	// middleware
	e.Use(mw.Recover())
	e.Use(echo.MiddlewareFunc(middleware.Logger(config.Name)))
	for _, middleware := range config.Middleware {
		e.Use(echo.MiddlewareFunc(middleware))
	}

	e.Get("/admin/health", makeHealthCheck(config.HealthCheck))
	// TODO admin health details
	e.Get("/admin/stats", statsHandler)
	e.Post("/admin/logging/:level", logHandler)

	s := &Server{
		e: e,
		c: config,
	}
	s.e.SetHTTPErrorHandler(s.httpErrorHandler)
	return s
}

// Start starts up the server and begins serving traffic
func (s *Server) Start() {
	s.g = &graceful.Server{
		Server: &http.Server{
			Addr:    s.c.Addr,
			Handler: s.e,
		},
		Timeout:        s.c.Timeout,
		BeforeShutdown: s.c.ShutdownFunc,
	}
	log.Infof("Starting server on %s", s.c.Addr)
	s.g.ListenAndServe()
}

// StartTest starts up the test server and begins serving traffic
func (s *Server) StartTest() string {
	s.t = httptest.NewServer(s.e)
	log.Infof("Started test server on %s", s.t.URL)
	return s.t.URL
}

// Stop gracefully shuts down the server
func (s *Server) Stop() {
	s.g.Stop(time.Second * 5)
}

// StopTest shuts down the test server
func (s *Server) StopTest() {
	s.t.Close()
}

// httpErrorHandler is identical to echo.DefaultHTTPErrorHandler except for using the right logger
func (s *Server) httpErrorHandler(err error, c *echo.Context) {
	if er, ok := err.(*errors.ErrorResponse); ok {
		ErrorResponse(c, er)
		return
	}
	code := http.StatusInternalServerError
	msg := http.StatusText(code)
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code()
		msg = he.Error()
	}
	if s.e.Debug() {
		msg = err.Error()
	}
	if !c.Response().Committed() {
		http.Error(c.Response(), msg, code)
	}
	log.Error(err)
}

// Echo returns the echo instance for this server
func (s *Server) Echo() *echo.Echo {
	return s.e
}

func makeHealthCheck(hcFunc func() error) func(c *echo.Context) error {
	return func(c *echo.Context) error {
		if hcFunc == nil {
			return echo.NewHTTPError(http.StatusNotImplemented, "Not Implemented")
		}

		err := hcFunc()
		if err != nil {
			echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		c.NoContent(http.StatusNoContent)
		return nil
	}
}

func statsHandler(c *echo.Context) error {
	c.Response().Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(c.Response(), "{\n")
	first := true
	expvar.Do(func(kv expvar.KeyValue) {
		if !first {
			fmt.Fprintf(c.Response(), ",\n")
		}
		first = false
		fmt.Fprintf(c.Response(), "%q: %s", kv.Key, kv.Value)
	})
	fmt.Fprintf(c.Response(), "\n}\n")
	return nil
}

func logHandler(c *echo.Context) error {
	logLevel := c.Param("level")
	err := log.SetLevel(logLevel)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}
