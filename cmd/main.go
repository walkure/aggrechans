package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"

	common "github.com/walkure/aggrechans"
)

var AGG_CHAN_ID = os.Getenv("AGGREGATE_CHANNEL_ID")

func createSlackSocketClient() (*slack.Client, *socketmode.Client, error) {
	appToken := os.Getenv("SLACK_APP_TOKEN")
	if appToken == "" {
		return nil, nil, errors.New("SLACK_APP_TOKEN must be set")
	}

	if !strings.HasPrefix(appToken, "xapp-") {
		return nil, nil, errors.New("SLACK_APP_TOKEN must have the prefix \"xapp-\"")
	}

	botToken := os.Getenv("SLACK_BOT_TOKEN")
	if botToken == "" {
		return nil, nil, errors.New("SLACK_BOT_TOKEN must be set")
	}

	if !strings.HasPrefix(botToken, "xoxb-") {
		return nil, nil, errors.New("SLACK_BOT_TOKEN must have the prefix \"xoxb-\"")
	}

	api := slack.New(
		botToken,
		slack.OptionAppLevelToken(appToken),
		//slack.OptionDebug(true),
		//slack.OptionLog(log.New(os.Stdout, "api: ", log.Lshortfile|log.LstdFlags)),
	)

	client := socketmode.New(
		api,
		//socketmode.OptionDebug(true),
		//socketmode.OptionLog(log.New(os.Stdout, "socketmode: ", log.Lshortfile|log.LstdFlags)),
	)

	return api, client, nil
}

func main() {
	api, client, err := createSlackSocketClient()
	if err != nil {
		fmt.Printf("cannot establish connectins:%v", err)
		os.Exit(-1)
	}

	chinfo, err := common.InitChanInfo(api)
	if err != nil {
		fmt.Printf("cannot init channel info:%v", err)
		os.Exit(-1)
	}

	uinfo, err := common.InitUserInfo(api)
	if err != nil {
		fmt.Printf("cannot init channel info:%v", err)
		os.Exit(-1)
	}

	go func() {
		for evt := range client.Events {
			switch evt.Type {
			case socketmode.EventTypeConnecting:
				fmt.Println("Connecting to Slack with Socket Mode...")
			case socketmode.EventTypeHello:
				fmt.Println("Hello from Slack with Socket Mode...")
			case socketmode.EventTypeConnectionError:
				fmt.Println("Connection failed. Retrying later...")
			case socketmode.EventTypeConnected:
				fmt.Println("Connected to Slack with Socket Mode.")
			case socketmode.EventTypeEventsAPI:
				eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
				if !ok {
					fmt.Printf("Ignored %+v\n", evt)
					continue
				}
				client.Ack(*evt.Request)
				switch eventsAPIEvent.Type {
				case slackevents.CallbackEvent:
					innerEvent := eventsAPIEvent.InnerEvent
					switch ev := innerEvent.Data.(type) {
					case *slackevents.MessageEvent:
						messageEventHandler(api, client, ev, chinfo, uinfo)
					case *slackevents.ChannelRenameEvent:
						chinfo.UpdateName(ev.Channel)
					case *slack.UserChangeEvent:
						uinfo.HandleUserChangeEvent(ev)
					case *slackevents.ChannelCreatedEvent:
						chinfo.HandleCreateEvent(ev.Channel)
					case *slackevents.ChannelUnarchiveEvent:
						name, err := chinfo.GetName(ev.Channel)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Failure handling unarchive channel(id=%s): %+v\n", ev.Channel, err)
						} else {
							fmt.Printf("Channel[%s](%s) unarchived\n", name, ev.Channel)
						}
					default:
						fmt.Printf("unsupported Callback Event received: %T\n", ev)
					}

				default:
					fmt.Printf("unsupported Events API event received: %s\n", eventsAPIEvent.Type)
				}
			default:
				fmt.Fprintf(os.Stderr, "Unexpected event type received: %s\n", evt.Type)
			}
		}
	}()

	client.Run()
}

func messageEventHandler(api *slack.Client, client *socketmode.Client, ev *slackevents.MessageEvent, ci *common.ChannelInfo, ui *common.UserInfo) {

	text := ev.Text
	uid := ev.User
	switch ev.SubType {
	default:
		fmt.Printf("ignore subtype:%s\n", ev.SubType)
		return
	case slack.MsgSubTypeMessageChanged:
		if ev.Message != nil {
			text = ev.Message.Text
			if ev.Message.Edited != nil {
				uid = ev.Message.Edited.User
			}
		}
	case slack.MsgSubTypeFileShare, slack.MsgSubTypeChannelTopic, slack.MsgSubTypeChannelPurpose, "":
	}

	if uid == "" {
		return
	}

	prof, err := ui.GetUserProfile(uid)
	if err != nil {
		fmt.Printf("cannot get user profile:%v\n", err)
		return
	}

	if prof.IsBots() {
		return
	}

	msgLink, err := ci.GetMessageLink(ev)
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
		msg, err = ui.ReplaceMentionUIDs(text)
		if err != nil {
			fmt.Printf("cannot resolve mentions:%v\n", err)
			return
		}
		msg = common.EscapeChannelCall(msg)
	}

	fullMsg := msgLink + " " + msg

	err = postMessage(api, prof, nil, disableUnfurlLink, fullMsg)
	if err != nil {
		fmt.Printf("postMessage err:%v\n", err)
	}
}

func postMessage(api *slack.Client, prof *common.UserProfile, blocks *[]slack.Block, disableUnfurlLink bool, msg string) error {

	options := []slack.MsgOption{slack.MsgOptionText(msg, false),
		slack.MsgOptionUsername(prof.Name),
		slack.MsgOptionIconURL(prof.Avatar)}

	if blocks != nil && len(*blocks) > 1 {
		options = append(options, slack.MsgOptionBlocks(*blocks...))
	}

	if disableUnfurlLink {
		options = append(options, slack.MsgOptionDisableLinkUnfurl())
	}

	_, _, err := common.PostMessageContext(context.TODO(), api, AGG_CHAN_ID, options...)
	return err
}
