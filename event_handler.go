package common

import (
	"context"
	"fmt"
	"os"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

type ExcludeChannelFunc func(chanName string) bool

func CallbackEventHandler(ctx context.Context, api *slack.Client, client *socketmode.Client, eventsAPIEvent slackevents.EventsAPIEvent,
	ci *ChannelInfo, ui *UserInfo, dstChannel string, excludeChFn ExcludeChannelFunc) {
	innerEvent := eventsAPIEvent.InnerEvent
	switch ev := innerEvent.Data.(type) {
	case *slackevents.MessageEvent:
		messageEventHandler(ctx, api, client, ev, ci, ui, dstChannel, excludeChFn)
	case *slackevents.ChannelRenameEvent:
		ci.UpdateName(ev.Channel)
	case *slack.UserChangeEvent:
		ui.HandleUserChangeEvent(ev)
	case *slackevents.ChannelCreatedEvent:
		ci.HandleCreateEvent(ev.Channel)
	case *slackevents.ChannelUnarchiveEvent:
		name, err := ci.GetName(ctx, ev.Channel)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failure handling unarchive channel(id=%s): %+v\n", ev.Channel, err)
		} else {
			fmt.Printf("Channel[%s](%s) unarchived\n", name, ev.Channel)
		}
	default:
		fmt.Printf("unsupported Callback Event received: %T\n", ev)
	}
}

func messageEventHandler(ctx context.Context, api *slack.Client, client *socketmode.Client, ev *slackevents.MessageEvent,
	ci *ChannelInfo, ui *UserInfo, dstChannel string, excludeChFn ExcludeChannelFunc) {

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

	if excludeChFn != nil {
		chanName, err := ci.GetName(ctx, ev.Channel)
		if err != nil {
			fmt.Printf("cannot resolve cnannel name(lookup):%v\n", err)
			return
		}

		if excludeChFn(chanName) {
			return
		}
	}

	msgLink, err := ci.GetMessageLink(ctx, ev)
	if err != nil {
		fmt.Printf("cannot resolve cnannel name(genLink):%v\n", err)
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
		msg = EscapeChannelCall(msg)
	}

	fullMsg := msgLink + " " + msg

	err = PostMessage(ctx, api, prof, nil, disableUnfurlLink, fullMsg, dstChannel)
	if err != nil {
		fmt.Printf("postMessage err:%v\n", err)
	}
}
