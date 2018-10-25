package session

import (
	"net/http"

	"github.com/viliproject/vili/util"
)

const (
	sessionCookie = "session"
)

// User is the user or robot making the request
type User struct {
	Email     string   `json:"email"`
	Username  string   `json:"username"`
	FirstName string   `json:"firstName"`
	LastName  string   `json:"lastName"`
	Groups    []string `json:"groups"`
}

var services []Service

// Service is an authentication service interface
type Service interface {
	Login(r *http.Request, w http.ResponseWriter, u *User) (bool, error)
	Logout(r *http.Request, w http.ResponseWriter) (bool, error)
	GetUser(r *http.Request) (*User, error)
}

// Login logs the user in with the auth service that was initialized
func Login(r *http.Request, w http.ResponseWriter, u *User) error {
	for _, service := range services {
		skip, err := service.Login(r, w, u)
		if err != nil {
			return err
		}
		if !skip {
			break
		}
	}
	return nil
}

// Logout logs the user in with the auth service that was initialized
func Logout(r *http.Request, w http.ResponseWriter) error {
	for _, service := range services {
		skip, err := service.Logout(r, w)
		if err != nil {
			return err
		}
		if !skip {
			break
		}
	}
	return nil
}

// GetUser logs the user in with the auth service that was initialized
func GetUser(r *http.Request) (*User, error) {
	for _, service := range services {
		user, err := service.GetUser(r)
		if user != nil || err != nil {
			return user, err
		}
	}
	return nil, nil
}

// helper functions
func getSessionCookie(r *http.Request) string {
	for _, cookie := range r.Cookies() {
		if cookie.Name == sessionCookie {
			return cookie.Value
		}
	}
	return ""
}

func newSessionID() string {
	return util.RandString(40)
}
