package common

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

type DestinationChannelFunc func(chanName string) string

func CallbackEventHandler(ctx context.Context, api *slack.Client, eventsAPIEvent slackevents.EventsAPIEvent,
	ci *ChannelInfo, ui *UserInfo, destChFn DestinationChannelFunc) error {
	innerEvent := eventsAPIEvent.InnerEvent
	switch ev := innerEvent.Data.(type) {
	case *slackevents.MessageEvent:
		return messageEventHandler(ctx, api, ev, ci, ui, destChFn)
	case *slackevents.ChannelRenameEvent:
		ci.UpdateName(ctx, ev.Channel)
	case *slack.UserChangeEvent:
		ui.HandleUserChangeEvent(ctx, ev)
	case *slackevents.ChannelCreatedEvent:
		ci.HandleCreateEvent(ctx, ev.Channel)
	case *slackevents.ChannelUnarchiveEvent:
		name, err := ci.GetName(ctx, ev.Channel)
		if err != nil {
			return fmt.Errorf("failure handling unarchive channel(id=%s):%w", ev.Channel, err)
		}
		fmt.Printf("Channel[%s](%s) unarchived\n", name, ev.Channel)
	default:
		return fmt.Errorf("unsupported Callback Event received: %T", ev)
	}
	return nil
}

func messageEventHandler(ctx context.Context, api *slack.Client, ev *slackevents.MessageEvent,
	ci *ChannelInfo, ui *UserInfo, destChFn DestinationChannelFunc) error {

	text := ev.Text
	uid := ev.User
	switch ev.SubType {
	case slack.MsgSubTypeBotMessage:
		return nil
	case slack.MsgSubTypeMessageChanged:
		if ev.Message != nil {
			text = ev.Message.Text
			if ev.Message.Edited != nil {
				uid = ev.Message.Edited.User
			}
		}
	case slack.MsgSubTypeFileShare, slack.MsgSubTypeChannelTopic, slack.MsgSubTypeChannelPurpose, slack.MsgSubTypeThreadBroadcast, "":
		// continue
	default:
		fmt.Printf("ignore subtype:%s\n", ev.SubType)
		return nil
	}

	if uid == "" {
		return nil
	}

	prof, err := ui.GetUserProfile(ctx, uid)
	if err != nil {
		return fmt.Errorf("cannot get user profile:%w", err)
	}

	if prof.IsBots() {
		return nil
	}

	chanName, err := ci.GetName(ctx, ev.Channel)
	if err != nil {
		return fmt.Errorf("cannot resolve cnannel name(lookup):%w", err)
	}

	dstChannel := destChFn(chanName)
	if dstChannel == "" {
		return nil
	}

	msgLink, err := ci.GetMessageLink(ctx, ev)
	if err != nil {
		return fmt.Errorf("cannot resolve cnannel name(genLink):%w", err)
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
			return fmt.Errorf("cannot resolve mentions:%w", err)
		}
		msg = EscapeChannelCall(msg)
	}

	fullMsg := msgLink + " " + msg

	err = PostMessage(ctx, api, prof, nil, disableUnfurlLink, fullMsg, dstChannel)
	if err != nil {
		return fmt.Errorf("postMessage err:%w", err)
	}

	return nil
}
