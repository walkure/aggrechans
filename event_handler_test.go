package common

import (
	"context"
	"testing"

	"github.com/slack-go/slack/slackevents"
	"github.com/stretchr/testify/assert"
)

func TestCallbackEventHandlerUserChange(t *testing.T) {
	ui := &UserInfo{name: map[string]*UserProfile{
		"U001": {Name: "before", Avatar: "https://example.com/before.png"},
	}}

	ev := slackevents.EventsAPIEvent{
		InnerEvent: slackevents.EventsAPIInnerEvent{
			Data: &slackevents.UserChangeEvent{
				User: slackevents.User{
					ID:        "U001",
					Name:      "after",
					IsBot:     true,
					IsAppUser: true,
					Profile:   slackevents.Profile{Image72: "https://example.com/after.png"},
				},
			},
		},
	}

	err := CallbackEventHandler(context.Background(), nil, ev, nil, ui, nil)

	assert.NoError(t, err)
	assert.Equal(t, "after", ui.name["U001"].Name)
	assert.Equal(t, "https://example.com/after.png", ui.name["U001"].Avatar)
	assert.True(t, ui.name["U001"].Bot)
	assert.True(t, ui.name["U001"].App)
}
