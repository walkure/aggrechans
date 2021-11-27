package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

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

	ctx := context.Background()
	wg := &sync.WaitGroup{}
	chinfo := &common.ChannelInfo{}
	uinfo := &common.UserInfo{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		chinfo, err = common.InitChanInfo(ctx, api)
		if err != nil {
			fmt.Printf("cannot init channel info:%v\n", err)
			os.Exit(-1)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		uinfo, err = common.InitUserInfo(ctx, api)
		if err != nil {
			fmt.Printf("cannot init user info:%v\n", err)
			os.Exit(-1)
		}
	}()

	wg.Wait()

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
					err := common.CallbackEventHandler(context.TODO(), api, client, eventsAPIEvent, chinfo, uinfo, AGG_CHAN_ID, nil)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error!:%+v\n", err)
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
