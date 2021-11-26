package main

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"

	common "github.com/walkure/aggrechans"
)

func messageEventHandler(ctx context.Context, api *slack.Client, client *socketmode.Client, ev *slackevents.MessageEvent, ci *common.ChannelInfo, ui *common.UserInfo) {

	text := ev.Text
	uid := ev.User
	switch ev.SubType {
	case slack.MsgSubTypeBotMessage:
		return
	case slack.MsgSubTypeMessageChanged:
		if ev.Message != nil {
			text = ev.Message.Text
			if ev.Message.Edited != nil {
				uid = ev.Message.Edited.User
			}
		}
	case slack.MsgSubTypeFileShare, slack.MsgSubTypeChannelTopic, slack.MsgSubTypeChannelPurpose, "":
		// continue
	default:
		fmt.Printf("ignore subtype:%s\n", ev.SubType)
		return
	}

	if uid == "" {
		return
	}

	prof, err := ui.GetUserProfile(ctx, uid)
	if err != nil {
		fmt.Printf("cannot get user profile:%v\n", err)
		return
	}

	if prof.IsBots() {
		return
	}

	msgLink, err := ci.GetMessageLink(ctx, ev)
	if err != nil {
		fmt.Printf("cannot resolve cnannel name:%v\n", err)
		return
	}

	msg := ""
	disableUnfurlLink := true
	switch ev.SubType {
	case slack.MsgSubTypeFileShare:
		msg = ""
		disableUnfurlLink = false
	default:
		msg, err = ui.ReplaceMentionUIDs(ctx, text)
		if err != nil {
			fmt.Printf("cannot resolve mentions:%v\n", err)
			return
		}
		msg = common.EscapeChannelCall(msg)
	}

	fullMsg := msgLink + " " + msg

	err = common.PostMessage(ctx, api, prof, nil, disableUnfurlLink, fullMsg, AGG_CHAN_ID)
	if err != nil {
		fmt.Printf("postMessage err:%v\n", err)
	}
}
