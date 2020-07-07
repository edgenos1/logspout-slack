package slack

import (
	"errors"
	"fmt"
	"os"
	"regexp"

	"github.com/gliderlabs/logspout/router"
	"github.com/slack-go/slack"
)

func init() {
	router.AdapterFactories.Register(NewSlackAdapter, "slack")
}

func getopt(name, dfault string) string {
	value := os.Getenv(name)
	if value == "" {
		value = dfault
	}
	return value
}

// NewSlackAdapter creates a Slack adapter.
func NewSlackAdapter(route *router.Route) (router.LogAdapter, error) {
	slackToken := getopt("SLACK_TOKEN", route.Options["slack_token"])
	slackUser := getopt("SLACK_USER", route.Options["slack_user"])
	slackUsername := getopt("SLACK_USERNAME", route.Options["slack_username"])
	slackChannel := getopt("SLACK_CHANNEL", route.Options["slack_channel"])
	messageFilter := getopt("SLACK_MESSAGE_FILTER", route.Options["slack_message_filter"])
	if messageFilter == "" {
		messageFilter = "*"
	}

	slackClient := slack.New(slackToken)
	if slackClient == nil {
		return nil, errors.New("unable to get slack client: " + route.Adapter)
	}

	return &SlackAdapter{
		slackClient:   slackClient,
		slackChannel:  slackChannel,
		slackUser:     slackUser,
		slackUsername: slackUsername,
		messageFilter: messageFilter,
		route:         route,
	}, nil
}

// SlackAdapter describes a Slack adapter
type SlackAdapter struct {
	slackClient   *slack.Client
	slackUser     string
	slackUsername string
	slackChannel  string
	messageFilter string
	route         *router.Route
}

// Stream implements the router.LogAdapter interface.
func (a *SlackAdapter) Stream(logstream chan *router.Message) {
	msgParams := slack.PostMessageParameters{
		User:     a.slackUser,
		Username: a.slackUsername,
	}
	msgOptionPostMsgParams := slack.MsgOptionPostMessageParameters(msgParams)
	fmt.Printf("%+v", a)
	for message := range logstream {
		if ok, _ := regexp.MatchString(a.messageFilter, message.Data); ok {
			a.slackClient.PostMessage(a.slackChannel, slack.MsgOptionText(message.Data, false), msgOptionPostMsgParams)
		}
	}
}
