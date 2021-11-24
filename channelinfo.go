package common

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

type ChannelInfo struct {
	name   map[string]string
	domain string
	api    *slack.Client
}

func InitChanInfo(api *slack.Client) (*ChannelInfo, error) {
	info := ChannelInfo{}
	info.name = make(map[string]string)
	info.api = api

	tinfo, err := api.GetTeamInfo()
	if err != nil {
		return nil, fmt.Errorf("err at team.info:%w", err)
	}
	info.domain = tinfo.Domain

	chans, err := getChannelList(context.Background(), api)
	if err != nil {
		return nil, fmt.Errorf("err at conversations.list:%w", err)
	}

	if len(chans) == 0 {
		return nil, errors.New("no conversation.list")
	}

	for _, ch := range chans {
		info.name[ch.ID] = ch.Name
	}
	//api.Debugf("loaded %d channels\n", len(info.name))
	fmt.Printf("loaded %d channels.team_domain=%s\n", len(info.name), info.domain)

	return &info, nil

}

func getChannelList(ctx context.Context, api *slack.Client) ([]slack.Channel, error) {
	req := slack.GetConversationsParameters{ExcludeArchived: true}
	var results []slack.Channel

	for {
		chans, cursor, err := api.GetConversationsContext(ctx, &req)

		if err == nil {
			results = append(results, chans...)
			if cursor == "" {
				return results, nil
			}
			req.Cursor = cursor
		} else if rateLimitedError, ok := err.(*slack.RateLimitedError); ok {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(rateLimitedError.RetryAfter):
				continue
			}
		} else {
			return nil, err
		}
	}
}

func (info *ChannelInfo) GetMessageLink(ev *slackevents.MessageEvent) (string, error) {
	name, err := info.GetName(ev.Channel)
	if err != nil {
		return "", fmt.Errorf("cannot convert channel id:%w", err)
	}

	return fmt.Sprintf("<%s|`#%s`>", info.GetMessageUri(ev), name), nil
}

func (info *ChannelInfo) GetMessageUri(ev *slackevents.MessageEvent) string {
	// cannot follow links at smartphone app. whils exists "." in  message id.
	mid := strings.Replace(ev.TimeStamp, ".", "", -1)
	return fmt.Sprintf("https://%s.slack.com/archives/%s/p%s", info.domain, ev.Channel, mid)
}

func (info *ChannelInfo) GetName(cid string) (string, error) {

	if name, ok := info.name[cid]; ok {
		return name, nil
	}

	cinfo, err := info.api.GetConversationInfo(cid, false)
	if err != nil {
		return "", fmt.Errorf("err at conversations.info(cid=%s):%w", cid, err)
	}

	info.name[cid] = cinfo.Name

	return info.name[cid], nil
}

func (info *ChannelInfo) UpdateName(chinfo slackevents.ChannelRenameInfo) {
	old, ok := info.name[chinfo.ID]
	if ok {
		fmt.Printf("rename channel(%s) %s -> %s\n", chinfo.ID, old, chinfo.Name)
	} else {
		fmt.Printf("rename channel(%s) ??? -> %s\n", chinfo.ID, chinfo.Name)
	}
	info.name[chinfo.ID] = chinfo.Name
}

func (info *ChannelInfo) HandleCreateEvent(chinfo slackevents.ChannelCreatedInfo) {
	info.name[chinfo.ID] = chinfo.Name
}
