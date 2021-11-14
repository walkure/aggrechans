package common

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/slack-go/slack"
)

func PostMessageContext(ctx context.Context, api *slack.Client, channelID string, options ...slack.MsgOption) (string, string, error) {
	for {
		respChannel, respTimestamp, err := api.PostMessageContext(ctx, channelID, options...)
		if err == nil {
			return respChannel, respTimestamp, nil
		} else if rateLimitedError, ok := err.(*slack.RateLimitedError); ok {
			select {
			case <-ctx.Done():
				return "", "", ctx.Err()
			case <-time.After(rateLimitedError.RetryAfter):
				continue
			}
		} else {
			return "", "", err
		}
	}
}

func EscapeChannelCall(orig string) string {
	origRune := []rune(orig)

	sb := strings.Builder{}
	tag := 0
	for i := 0; i < len(origRune); i++ {
		if origRune[i] == '<' {
			tag = 1
			sb.WriteRune('<')
		} else if origRune[i] == '!' && tag == 1 {
			sbb := strings.Builder{}
			tagClosed := false
			for j := i; j < len(origRune); j++ {
				if origRune[j] == '>' {
					tagClosed = true
					break
				} else {
					sbb.WriteRune(origRune[j])
				}
			}
			if !tagClosed {
				sb.WriteRune('！')
				tag = 0
				continue
			}
			sbbs := sbb.String()
			fmt.Println(sbbs)
			if strings.HasPrefix(sbbs, "!subteam^") && len(sbbs) > 22 {
				sb.WriteRune('＠')
				i += 21
			} else if sbbs == "!channel" {
				sb.WriteString("@channel")
				i += 7
			} else if sbbs == "!everyone" {
				sb.WriteString("@everyone")
				i += 8
			} else if sbbs == "!here" {
				sb.WriteString("@here")
				i += 4
			} else if sbbs == "!group" {
				sb.WriteString("@group")
				i += 5
			} else {
				sb.WriteRune('！')
			}
			tag = 0
		} else {
			sb.WriteRune(origRune[i])
			tag = 0
		}
	}
	return sb.String()
}