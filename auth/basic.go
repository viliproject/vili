package auth

import (
	"net/http"
	"strings"

	"github.com/labstack/echo"
	"github.com/viliproject/vili/server"
	"github.com/viliproject/vili/session"
	"golang.org/x/crypto/bcrypt"
)

// BasicConfig is the configuration for the BasicAuthService
type BasicConfig struct {
	Users []string
}

// BasicAuthService is the auth service that uses Basic to authenticate users
type BasicAuthService struct {
	config      *BasicConfig
	credentials map[string]string
}

// InitBasicAuthService creates a new instance of BasicAuthService from the given
// config and sets it as the default auth service
func InitBasicAuthService(config *BasicConfig) error {
	credentials := map[string]string{}
	for _, user := range config.Users {
		splitUser := strings.SplitN(user, ":", 2)
		if len(splitUser) == 2 {
			credentials[splitUser[0]] = splitUser[1]
		}
	}
	service = &BasicAuthService{
		config:      config,
		credentials: credentials,
	}
	return nil
}

// AddHandlers implements the Service interface
func (s *BasicAuthService) AddHandlers(srv *server.Server) {
	srv.Echo().GET("/login", s.loginHandler)
}

// Cleanup implements the Service interface
func (s *BasicAuthService) Cleanup() {
}

func (s *BasicAuthService) loginHandler(c echo.Context) error {
	r := c.Request()
	w := c.Response()

	username, password, ok := r.BasicAuth()
	// check password
	if !ok || (password != "" && bcrypt.CompareHashAndPassword([]byte(s.credentials[username]), []byte(password)) != nil) {
		// ask for http login
		w.Header().Set("WWW-Authenticate", `Basic realm="vili"`)
		return echo.NewHTTPError(http.StatusUnauthorized)
	}

	redirectURI := c.Request().URL.Query().Get("redirect")
	if redirectURI == "" {
		redirectURI = "/"
	}

	user := &session.User{
		Email:     username + "@basic",
		Username:  username,
		FirstName: username,
		LastName:  "",
	}

	err := session.Login(r, c.Response(), user)
	if err != nil {
		return err
	}

	return c.Redirect(http.StatusFound, redirectURI)
}
