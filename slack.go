package slack

import (
	"os"
	"regexp"
	"fmt"
	"strings"
	 "bytes"
	"text/template"
	
	"github.com/gliderlabs/logspout/router"
	"github.com/slack-go/slack"
)

func init() {
	router.AdapterFactories.Register(NewSlackAdapter, "slack")
}

// Get options from env var or default config
func getopt(name, dfault string) string {
	value := os.Getenv(name)
	if value == "" {
		value = dfault
	}
	return value
}

// Get env var as map
func getenv() map[string]string {
        items := make(map[string]string)
        for _, item := range os.Environ() {
		splits := strings.Split(item, "=")
		key := splits[0]
		value := splits[1]
            	items[key] = value
        }
        return items
}

// NewSlackAdapter creates a Slack adapter.
func NewSlackAdapter(route *router.Route) (router.LogAdapter, error) {
	slackWebhook := getopt("SLACK_WEBHOOK_URL", "")
	if !strings.HasPrefix(slackWebhook, "https://") {
		slackWebhook = "https://" + route.Address + slackWebhook
	}
	messageFilter := getopt("SLACK_MESSAGE_FILTER", ".*?")
	titleTemplateExpression := getopt("SLACK_TITLE_TEMPLATE", "{{ .Message.Container.Name}}")
	messageTemplateExpression := getopt("SLACK_MESSAGE_TEMPLATE", "{{ .Message.Data}}")
	linkTemplateExpression := getopt("SLACK_LINK_TEMPLATE", "")
	colorTemplateExpression := getopt("SLACK_COLOR_TEMPLATE", "danger")
	titleTemplate, errTitle := template.New("title").Parse(titleTemplateExpression)
	if errTitle != nil {
		panic(errTitle)
	}
	messageTemplate, errMessage := template.New("message").Parse(messageTemplateExpression)
	if errMessage != nil {
		panic(errMessage)
	}
	linkTemplate, errLink := template.New("link").Parse(linkTemplateExpression)
	if errLink != nil {
		panic(errLink)
	}
	colorTemplate, errColor := template.New("color").Parse(colorTemplateExpression)
	if errColor != nil {
		panic(errColor)
	}

	fmt.Printf("Creating Slack adapter on webhook %v with filter %v\n", slackWebhook, messageFilter)
	return &SlackAdapter{
		slackWebhook:   	slackWebhook,
		messageFilter: 		messageFilter,
		titleTemplate: 		titleTemplate,
		messageTemplate: 	messageTemplate,
		linkTemplate: 		linkTemplate,
		colorTemplate: 		colorTemplate,
		route:         		route,
	}, nil
}

// SlackAdapter describes a Slack adapter
type SlackAdapter struct {
	slackWebhook  	string
	messageFilter 	string
	titleTemplate	*template.Template
	messageTemplate	*template.Template
	linkTemplate	*template.Template
	colorTemplate	*template.Template
	route         	*router.Route
}

type Context struct {
	Message 	*router.Message
	Env     	*map[string]string
}

// Stream implements the router.LogAdapter interface.
func (a *SlackAdapter) Stream(logstream chan *router.Message) {
	env := getenv()
	for message := range logstream {
		if message.Data == "" {
			continue
		}
		fmt.Printf("Filtering message for slack: %+v\n", message.Data)
		if ok, _ := regexp.MatchString(a.messageFilter, message.Data); ok {
			fmt.Printf("Sending slack message: %+v\n", message.Data)
			context := Context{
				Message: message,
				Env: &env,
			}
			var buffer bytes.Buffer
			color := "danger"
			if errColor := a.colorTemplate.Execute(&buffer, context); errColor == nil {
			    color = buffer.String()
			}
			buffer.Reset()
			title := "[ALERT] Logspout"
			if errTitle := a.titleTemplate.Execute(&buffer, context); errTitle == nil {
			    title = buffer.String()
			}
			buffer.Reset()
			link := ""
			if errLink := a.linkTemplate.Execute(&buffer, context); errLink == nil {
			    link = buffer.String()
			}
			buffer.Reset()
			message := message.Data
			if errMesssage := a.messageTemplate.Execute(&buffer, context); errMesssage == nil {
			    message = buffer.String()
			}
			
			attachment := slack.Attachment{
				Color:		color,
				Title:		title,
				TitleLink:	link,
				Text:		message,
			}
			msg := slack.WebhookMessage{
				Attachments: []slack.Attachment{attachment},
			}
			slack.PostWebhook(a.slackWebhook, &msg)
		}
	}
}
