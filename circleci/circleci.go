package circleci

import (
	"github.com/airware/vili/errors"

	"github.com/airware/vili/log"
	"github.com/jszwedko/go-circleci"
)

var config *Config
var client *circleci.Client

// Config is the circle configuration
type Config struct {
	Token string
}

// Init Initializes the circle ci client
func Init(c *Config) error {
	config = c
	if c.Token == "" {
		return errors.BadRequest("Missing circle ci token")
	}
	client = &circleci.Client{Token: c.Token}
	c.Printf("Initialized ci client for %s", c)
	return nil
}

// Printf is the minimal logging method from Logger interface
func (c *Config) Printf(format string, args ...interface{}) {
	log.Infof(format, args)
}

// CircleBuild runs a build on circle ci for the defined branch
func CircleBuild(account, repo, branch string, buildParameters map[string]string) (*circleci.Build, error) {
	return client.ParameterizedBuild(account, repo, branch, buildParameters)
}
