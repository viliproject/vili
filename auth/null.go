package auth

import (
	"net/http"

	"github.com/airware/vili/server"
	"github.com/airware/vili/session"
	"github.com/labstack/echo"
)

// Null implements a the auth Service interface and always auths
type Null struct{}

// InitNullAuthService sets the auth service to null
func InitNullAuthService() error {
	service = &Null{}
	return nil
}

// AddHandlers implements the Service interface
func (s *Null) AddHandlers(srv *server.Server) {
	srv.Echo().GET("/login", s.loginHandler)
}

// Cleanup is a noop with the null handler
func (s *Null) Cleanup() {
	return
}

func (s *Null) loginHandler(c echo.Context) error {
	err := session.Login(c.Request(), c.Response(), &session.User{
		Email:     "dev@dev.local",
		Username:  "dev@dev.local",
		FirstName: "nullauth",
		LastName:  "nullauth",
	})
	if err != nil {
		return err
	}
	return c.Redirect(http.StatusFound, "/")
}
