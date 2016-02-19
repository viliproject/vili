package slack

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/airware/vili/log"
	"github.com/airware/vili/util"
	"github.com/nlopes/slack"
)

// WaitGroup is the wait group to synchronize slack rtm shutdowns
var WaitGroup sync.WaitGroup

// Exiting is a flag indicating that the server is exiting
var Exiting = false

var config *Config
var client *slack.Client

// Config is the slack configuration
type Config struct {
	Token           string
	Channel         string
	Username        string
	Emoji           string
	DeployUsernames []string
}

// Init initializes the slack client
func Init(c *Config) error {
	config = c
	client = slack.New(c.Token)
	return nil
}

// PostLogMessage posts a formatted log message to slack
func PostLogMessage(message, level string) error {
	if client == nil {
		return nil
	}
	color := "#36a64f"
	switch level {
	case "error":
		color = "#ff3300"
	case "warn":
		color = "#ffaa00"
	}

	_, _, err := client.PostMessage(config.Channel, "", slack.PostMessageParameters{
		Username:  config.Username,
		IconEmoji: config.Emoji,
		Attachments: []slack.Attachment{
			slack.Attachment{
				Color:      color,
				Text:       message,
				MarkdownIn: []string{"text"},
			},
		},
	})
	return err
}

// ListenForMentions opens an RTM connection to slack and listens for any mentions of the configured user
// in the configured channel
func ListenForMentions(mentions chan<- *Mention) {
	var waitGroup sync.WaitGroup
	var botID string
	var channelID string
	deployUsers := map[string]string{}
	failed := false

	channelName := strings.TrimLeft(config.Channel, "#")

	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		// get a list of user ids with deploy permissions
		users, err := client.GetUsers()
		if err != nil {
			log.Error(err)
			failed = true
			return
		}
		for _, user := range users {
			if user.Name == config.Username {
				botID = user.ID
			} else if util.Contains(config.DeployUsernames, user.Name) {
				deployUsers[user.ID] = user.Name
			}
		}
	}()

	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		// look for the channel id
		channels, err := client.GetChannels(true)
		if err != nil {
			log.Error(err)
			failed = true
			return
		}
		for _, channel := range channels {
			if channel.Name == channelName {
				channelID = channel.ID
				break
			}
		}
	}()

	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		// look for the channel id
		groups, err := client.GetGroups(true)
		if err != nil {
			log.Error(err)
			failed = true
			return
		}
		for _, group := range groups {
			if group.Name == channelName {
				channelID = group.ID
				break
			}
		}
	}()

	waitGroup.Wait()
	if failed {
		return
	}

	if botID == "" {
		log.Errorf("User %s not found", config.Username)
		return
	}
	if channelID == "" {
		log.Errorf("Channel %s not found", config.Channel)
		return
	}
	botMention := fmt.Sprintf("<@%s>: ", botID)
	botNameMention := fmt.Sprintf("<@%s>: ", config.Username)

	rtm := client.NewRTM()
	go rtm.ManageConnection()

	WaitGroup.Add(1)
	go func() {
	IncomingEvents:
		for event := range rtm.IncomingEvents {
			switch ev := event.Data.(type) {

			case *slack.MessageEvent:
				username := deployUsers[ev.User]
				if username != "" && ev.Channel == channelID {
					log.Debug(ev.User)
					log.Debug(ev.Text)
				}
				if username != "" && ev.Channel == channelID &&
					(strings.HasPrefix(ev.Text, botMention) || strings.HasPrefix(ev.Text, botNameMention)) {
					mentions <- &Mention{
						Timestamp: ev.Timestamp,
						Text:      strings.TrimPrefix(strings.TrimPrefix(ev.Text, botMention), botNameMention),
						Username:  username,
					}
				}

			case *slack.RTMError:
				log.Error(ev.Error())

			case *slack.DisconnectedEvent:
				log.Info("RTM disconnected")
				break IncomingEvents

			default:
				// ignore other events
			}
		}
		WaitGroup.Done()
	}()

	log.Info("Started slack rtm")
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for !Exiting {
		<-ticker.C
	}
	log.Info("Disconnecting slack rtm")
	err := rtm.Disconnect()
	if err != nil {
		log.Error(err)
	}
	close(mentions)
}

// Mention is a mention of the vili bot from the configured channel
type Mention struct {
	Timestamp string // this is guaranteed to be unique per channel
	Text      string
	Username  string
}
