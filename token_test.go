package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadSlackToken(t *testing.T) {
	tests := []struct {
		name    string
		bot     string
		user    string
		want    SlackToken
		wantErr bool
	}{
		{name: "bot", bot: "xoxb-foo", want: SlackToken{Token: "xoxb-foo"}},
		{name: "user", user: "xoxp-foo", want: SlackToken{Token: "xoxp-foo", IsUser: true}},
		{name: "none", wantErr: true},
		{name: "both", bot: "xoxb-foo", user: "xoxp-foo", wantErr: true},
		{name: "bad bot prefix", bot: "xoxp-foo", wantErr: true},
		{name: "bad user prefix", user: "xoxb-foo", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SLACK_BOT_TOKEN", tt.bot)
			t.Setenv("SLACK_USER_TOKEN", tt.user)

			got, err := LoadSlackToken()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
