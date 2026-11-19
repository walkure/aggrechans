package common

import (
	"context"
	"testing"

	"github.com/slack-go/slack/slackevents"
	"github.com/stretchr/testify/assert"
)

func TestUserChangeEvent(t *testing.T) {
	body := `{
	"type": "event_callback",
	"event": {
		"type": "user_change",
		"user": {
			"id": "U123",
			"name": "alice_renamed",
			"is_bot": false,
			"is_app_user": false,
			"profile": {"image_72": "https://example.com/new.png"}
		},
		"cache_ts": 1234567890,
		"event_ts": "1234567890.123456"
	}
}`
	ev, err := slackevents.ParseEvent([]byte(body), slackevents.OptionNoVerifyToken())
	assert.NoError(t, err)

	ui := CreateUserInfo(nil, nil)
	ui.name["U123"] = &UserProfile{Name: "alice", Avatar: "https://example.com/old.png"}

	err = CallbackEventHandler(context.Background(), nil, ev, nil, ui, nil)
	assert.NoError(t, err)

	prof, ok := ui.lookupUserInfo(context.Background(), "U123")
	assert.True(t, ok)
	assert.Equal(t, &UserProfile{Name: "alice_renamed", Avatar: "https://example.com/new.png"}, prof)
}
