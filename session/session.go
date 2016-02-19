package session

import (
	"net/http"

	"github.com/airware/vili/util"
)

const (
	sessionCookie = "session"
)

// User is the user or robot making the request
type User struct {
	Email     string `json:"email"`
	Username  string `json:"username"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

var service Service

// Service is an authentication service interface
type Service interface {
	Login(r *http.Request, w http.ResponseWriter, u *User) error
	Logout(r *http.Request, w http.ResponseWriter) error
	GetUser(r *http.Request) (*User, error)
}

// Login logs the user in with the auth service that was initialized
func Login(r *http.Request, w http.ResponseWriter, u *User) error {
	return service.Login(r, w, u)
}

// Logout logs the user in with the auth service that was initialized
func Logout(r *http.Request, w http.ResponseWriter) error {
	return service.Logout(r, w)
}

// GetUser logs the user in with the auth service that was initialized
func GetUser(r *http.Request) (*User, error) {
	return service.GetUser(r)
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
