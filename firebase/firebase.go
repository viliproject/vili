package firebase

import (
	"strings"

	"github.com/CloudCom/firego"
	"github.com/viliproject/vili/log"
)

var (
	config   *Config
	database *firego.Firebase

	// ExitingChan is a flag indicating that the server is exiting
	ExitingChan = make(chan struct{})
)

// Config is the firebase configuration
type Config struct {
	URL    string
	Secret string
}

// Init initializes the firebase connection
func Init(c *Config) error {
	config = c

	database = firego.New(config.URL)
	database.Auth(config.Secret)

	return nil
}

// Database returns the initialized database connection
func Database() *firego.Firebase {
	return database
}

// Watch listens for changes on the given path and sends the events to the given chan
func Watch(path string, eventsChan chan Event) error {
	url := strings.TrimSuffix(config.URL, "/") + path
	log.WithField("url", url).Debug("listening to firebase path")
	db := firego.New(url)
	db.Auth(config.Secret)

	c := make(chan firego.Event)
	if err := db.Watch(c); err != nil {
		return err
	}

	watchChan := make(chan struct{})
	go func() {
		for event := range c {
			eventsChan <- Event(event)
		}
		close(watchChan)
	}()

	// wait until close is called on either watchChan or ExitingChan
	select {
	case <-watchChan:
		break
	case <-ExitingChan:
		break
	}

	db.StopWatching()
	return nil
}

// Event represents a notification received when watching a
// firebase reference
type Event struct {
	// Type of event that was received
	Type string
	// Path to the data that changed
	Path string
	// Data that changed
	Data interface{}
}
