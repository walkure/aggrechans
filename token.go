package common

import (
	"errors"
	"os"
	"strings"
)

// SlackToken is the token used for Web API calls.
type SlackToken struct {
	Token string
	// IsUser is true when Token is a user token (xoxp-).
	// User tokens cannot customize the username/icon of posted messages.
	IsUser bool
}

// LoadSlackToken loads either SLACK_BOT_TOKEN(xoxb-) or SLACK_USER_TOKEN(xoxp-).
func LoadSlackToken() (SlackToken, error) {
	botToken := os.Getenv("SLACK_BOT_TOKEN")
	userToken := os.Getenv("SLACK_USER_TOKEN")

	switch {
	case botToken != "" && userToken != "":
		return SlackToken{}, errors.New("set only one of SLACK_BOT_TOKEN or SLACK_USER_TOKEN")
	case botToken != "":
		if !strings.HasPrefix(botToken, "xoxb-") {
			return SlackToken{}, errors.New("SLACK_BOT_TOKEN must have the prefix \"xoxb-\"")
		}
		return SlackToken{Token: botToken}, nil
	case userToken != "":
		if !strings.HasPrefix(userToken, "xoxp-") {
			return SlackToken{}, errors.New("SLACK_USER_TOKEN must have the prefix \"xoxp-\"")
		}
		return SlackToken{Token: userToken, IsUser: true}, nil
	default:
		return SlackToken{}, errors.New("SLACK_BOT_TOKEN or SLACK_USER_TOKEN must be set")
	}
}
