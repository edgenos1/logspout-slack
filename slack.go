package slack

import (
	"os"
	"regexp"
	"fmt"

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
	slackWebhook := getopt("SLACK_WEBHOOK_URL", route.Address)
	messageFilter := getopt("SLACK_MESSAGE_FILTER", route.Options["slack_message_filter"])
	if messageFilter == "" {
		messageFilter = "*"
	}

	fmt.Printf("Creating Slack adapter with filter: %v\n", messageFilter)
	return &SlackAdapter{
		slackWebhook:   slackWebhook,
		messageFilter: 	messageFilter,
		route:         	route,
	}, nil
}

// SlackAdapter describes a Slack adapter
type SlackAdapter struct {
	slackWebhook  string
	messageFilter string
	route         *router.Route
}

// Stream implements the router.LogAdapter interface.
func (a *SlackAdapter) Stream(logstream chan *router.Message) {
	for message := range logstream {
		fmt.Printf("Filtering message for slack: %+v\n", message.Data)
		if ok, _ := regexp.MatchString(a.messageFilter, message.Data); ok {
			fmt.Printf("Sending slack message: %+v\n", message.Data)
			msg := slack.WebhookMessage{
				Text:     message.Data,
			}
			slack.PostWebhook(a.slackWebhook, &msg)
		}
	}
}
