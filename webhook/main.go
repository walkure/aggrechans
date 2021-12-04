package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	common "github.com/walkure/aggrechans"
)

var AGG_CHAN_ID = os.Getenv("AGGREGATE_CHANNEL_ID")

func createSlackClient() (*slack.Client, error) {
	botToken := os.Getenv("SLACK_BOT_TOKEN")
	if botToken == "" {
		return nil, errors.New("SLACK_BOT_TOKEN must be set")
	}

	if !strings.HasPrefix(botToken, "xoxb-") {
		return nil, errors.New("SLACK_BOT_TOKEN must have the prefix \"xoxb-\"")
	}

	return slack.New(
		botToken,
		//slack.OptionDebug(true),
		//slack.OptionLog(log.New(os.Stdout, "api: ", log.Lshortfile|log.LstdFlags)),
	), nil
}

func main() {
	signingSecret := os.Getenv("SLACK_SIGNING_SECRET")
	api, err := createSlackClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot load slack config:%+v\n", err)
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	opt := common.LoadRedisConfig()
	if opt == nil {
		fmt.Fprintln(os.Stderr, "cannot load redis config.")
		return
	}
	redis := redis.NewClient(opt)
	uinfo := common.CreateUserInfo(api, redis)
	chinfo, _ := common.CreateChanInfo(ctx, api, redis)

	dispatcher, err := common.NewDispatcher()
	if err != nil {
		fmt.Printf("cannot load dispatch info:%v", err)
		os.Exit(-1)
	}
	fmt.Println(dispatcher.Rules())

	http.HandleFunc("/events-endpoint", func(w http.ResponseWriter, r *http.Request) {
		body, err := loadRequest(w, r, signingSecret)
		if err != nil {
			return
		}
		eventsAPIEvent, err := slackevents.ParseEvent(body, slackevents.OptionNoVerifyToken())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		switch eventsAPIEvent.Type {
		case slackevents.URLVerification:
			var r *slackevents.ChallengeResponse
			err := json.Unmarshal([]byte(body), &r)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			fmt.Printf("Respond to Slack challenge")
			w.Header().Set("Content-Type", "text")
			w.Write([]byte(r.Challenge))
		case slackevents.CallbackEvent:
			go func() {
				err := common.CallbackEventHandler(r.Context(), api, eventsAPIEvent, chinfo, uinfo, dispatcher.Dispatch)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error!:%+v\n", err)
				}
			}()
		}

	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	serv := &http.Server{
		Addr:    ":" + port,
		Handler: nil,
	}

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		fmt.Println("[INFO] SIGINT received. byebye~")
		serv.Shutdown(ctx)
	}()

	fmt.Printf("[INFO] Server listening at %s\n", port)
	fmt.Fprintf(os.Stderr, "[FATAL] server shutdown. %+v\n", serv.ListenAndServe())
}

func loadRequest(w http.ResponseWriter, r *http.Request, signingSecret string) (json.RawMessage, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil, fmt.Errorf("cannot read body:%w", err)
	}
	sv, err := slack.NewSecretsVerifier(r.Header, signingSecret)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil, fmt.Errorf("cannot init secret verifier:%w", err)
	}
	if _, err := sv.Write(body); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return nil, fmt.Errorf("cannot write body:%w", err)
	}
	if err := sv.Ensure(); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return nil, fmt.Errorf("signature not match:%w", err)
	}

	return json.RawMessage(body), nil

}
