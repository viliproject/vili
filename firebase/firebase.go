package firebase

import (
	"github.com/CloudCom/fireauth"
	"github.com/CloudCom/firego"
	"github.com/airware/vili/session"
)

var database *firego.Firebase
var generator *fireauth.Generator

// Config is the firebase configuration
type Config struct {
	URL    string
	Secret string
}

// Init initializes the firebase connection
func Init(c *Config) error {
	database = firego.New(c.URL)
	database.Auth(c.Secret)

	generator = fireauth.New(c.Secret)

	return nil
}

// Database returns the initialized database connection
func Database() *firego.Firebase {
	return database
}

var authOptions = &fireauth.Option{}

// NewToken returns a new firebase token for the user
func NewToken(user *session.User) (string, error) {
	return generator.CreateToken(fireauth.Data{
		"uid":      user.Email,
		"username": user.Username,
		"role": map[string]bool{
			"dev": true, // TODO get these from Okta
			"qa":  true,
		},
	}, authOptions)
}
