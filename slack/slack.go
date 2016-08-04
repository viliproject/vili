package slack

import (
	"fmt"
	"regexp"
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

var slackLinkRegexp = regexp.MustCompile(`<.+?\|.+?>`)

// Config is the slack configuration
type Config struct {
	Token           string
	Channel         string
	Username        string
	Emoji           string
	DeployUsernames *util.StringSet
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

// MentionsRegexp returns a regexp matching messages which mention the configured user
func MentionsRegexp() (*regexp.Regexp, error) {
	var botID string
	users, err := client.GetUsers()
	if err != nil {
		return nil, err
	}
	for _, user := range users {
		if user.Name == config.Username {
			botID = user.ID
			break
		}
	}
	if botID == "" {
		return nil, fmt.Errorf("User %s not found", config.Username)
	}

	return regexp.Compile(fmt.Sprintf("^<@(?:%s|%s)>: (.*)", botID, config.Username))

}

// ListenForMessages opens an RTM connection to slack and listens for any messages
// in the configured channel which matches any of the regexps and sends the message
// to the mapped channel
func ListenForMessages(messageMap map[*regexp.Regexp]chan<- *Message) {
	var waitGroup sync.WaitGroup
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
			if config.DeployUsernames.Contains(user.Name) {
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

	if channelID == "" {
		log.Errorf("Channel %s not found", config.Channel)
		return
	}

	rtm := client.NewRTM()
	go rtm.ManageConnection()

	WaitGroup.Add(1)
	go func() {
	IncomingEvents:
		for event := range rtm.IncomingEvents {
			switch ev := event.Data.(type) {

			case *slack.MessageEvent:
				username, ok := deployUsers[ev.User]
				if !ok {
					continue
				}
				if ev.Channel == channelID {
					log.Debug(ev.User)
					log.Debug(ev.Text)
					for re, ch := range messageMap {
						if match := re.FindStringSubmatch(ev.Text); len(match) > 0 {
							ch <- &Message{
								Timestamp: ev.Timestamp,
								Text:      match[0],
								Matches:   match[1:],
								Username:  username,
							}
						} else {
							if match := re.FindStringSubmatch(removeLinks(ev.Text)); len(match) > 0 {
								ch <- &Message{
									Timestamp: ev.Timestamp,
									Text:      match[0],
									Matches:   match[1:],
									Username:  username,
								}
							}
						}
					}
				}

			case *slack.RTMError:
				log.Error(ev.Error())

			case *slack.ConnectedEvent:
				log.Infof("RTM Connected, Connection count: %d", ev.ConnectionCount)

			case *slack.DisconnectedEvent:
				log.Infof("RTM disconnected, Intentional: %v", ev.Intentional)
				if ev.Intentional {
					break IncomingEvents
				}

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
	for _, ch := range messageMap {
		close(ch)
	}
}

func removeLinks(message string) string {
	return slackLinkRegexp.ReplaceAllStringFunc(message, func(match string) string {
		sepIndex := strings.LastIndex(match, "|")
		return match[sepIndex+1 : len(match)-1]
	})
}

// Message is a slack message from the configured channel
type Message struct {
	Timestamp string // this is guaranteed to be unique per channel
	Text      string
	Matches   []string
	Username  string
}
