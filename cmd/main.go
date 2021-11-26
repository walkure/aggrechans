package main

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"

	common "github.com/walkure/aggrechans"
)

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
			fmt.Printf("cannot init channel info:%v", err)
			os.Exit(-1)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		uinfo, err = common.InitUserInfo(ctx, api)
		if err != nil {
			fmt.Printf("cannot init channel info:%v", err)
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
					innerEvent := eventsAPIEvent.InnerEvent
					switch ev := innerEvent.Data.(type) {
					case *slackevents.MessageEvent:
						messageEventHandler(context.TODO(), api, client, ev, chinfo, uinfo)
					case *slackevents.ChannelRenameEvent:
						chinfo.UpdateName(ev.Channel)
					case *slack.UserChangeEvent:
						uinfo.HandleUserChangeEvent(ev)
					case *slackevents.ChannelCreatedEvent:
						chinfo.HandleCreateEvent(ev.Channel)
					case *slackevents.ChannelUnarchiveEvent:
						name, err := chinfo.GetName(context.TODO(), ev.Channel)
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
