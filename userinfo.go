package common

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/slack-go/slack"
)

type UserInfo struct {
	name map[string]*UserProfile
	api  *slack.Client
	mu   sync.Mutex
}

type UserProfile struct {
	Name   string
	Avatar string
	bot    bool
	app    bool
}

func CreateUserInfo(api *slack.Client) *UserInfo {
	info := UserInfo{}
	info.name = make(map[string]*UserProfile)
	info.api = api

	return &info
}

func InitUserInfo(ctx context.Context, api *slack.Client) (*UserInfo, error) {
	info := CreateUserInfo(api)

	users, err := api.GetUsersContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("err at users.list:%w", err)
	}
	for _, user := range users {
		info.setUserInfo(&user)
	}
	//api.Debugf("loaded %d users\n", len(info.name))
	fmt.Printf("loaded %d users\n", len(info.name))

	return info, nil
}

func (prof *UserProfile) IsBots() bool {
	return prof.bot || prof.app
}

func (info *UserInfo) GetUserProfile(ctx context.Context, uid string) (*UserProfile, error) {
	if prof, ok := info.lookupUserInfo(uid); ok {
		return prof, nil
	}

	if strings.HasPrefix(uid, "B") {
		bot, err := info.api.GetBotInfoContext(ctx, uid)
		if err != nil {
			return nil, fmt.Errorf("err at bot.info(uid=%s):%w", uid, err)
		}

		return info.setBotInfo(bot), nil
	}

	user, err := info.api.GetUserInfoContext(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("err at users.info(uid=%s):%w", uid, err)
	}

	return info.setUserInfo(user), nil
}

func (info *UserInfo) HandleUserChangeEvent(ev *slack.UserChangeEvent) {
	info.setUserInfo(&ev.User)
}

func (info *UserInfo) lookupUserInfo(uid string) (*UserProfile, bool) {
	info.mu.Lock()
	defer info.mu.Unlock()

	prof, ok := info.name[uid]
	return prof, ok
}

func (info *UserInfo) setUserInfo(user *slack.User) *UserProfile {
	prof := &UserProfile{
		Name:   user.Name,
		Avatar: user.Profile.Image72,
		bot:    user.IsBot || user.ID == "USLACKBOT",
		app:    user.IsAppUser,
	}

	info.mu.Lock()
	defer info.mu.Unlock()

	info.name[user.ID] = prof

	return prof
}

func (info *UserInfo) setBotInfo(bot *slack.Bot) *UserProfile {
	prof := &UserProfile{
		Name:   bot.Name,
		Avatar: bot.Icons.Image72,
		bot:    true,
		app:    true,
	}

	info.mu.Lock()
	defer info.mu.Unlock()

	info.name[bot.ID] = prof

	return prof
}

func (info *UserInfo) ReplaceMentionUIDs(ctx context.Context, orig string) (string, error) {

	uids := extractUids(orig)

	oldnews := []string{}
	for _, uid := range uids {
		prof, err := info.GetUserProfile(ctx, uid)
		if err != nil {
			return "", fmt.Errorf("error replacing uids:%w", err)
		}
		oldnews = append(oldnews, "<@"+uid+">", "<＠"+prof.Name+">")
	}

	reps := strings.NewReplacer(oldnews...)
	return reps.Replace(orig), nil
}

func extractUids(msg string) []string {
	msgRune := []rune(msg)

	uids := []string{}
	tag := 0
	for i := 0; i < len(msgRune); i++ {
		if msgRune[i] == '<' {
			tag = 1
		} else if msgRune[i] == '@' && tag == 1 {
			tag = 2
		} else if msgRune[i] == 'U' && tag == 2 {
			sb := strings.Builder{}
			for j := i; j < len(msgRune); j++ {
				if msgRune[j] == '>' {
					uids = append(uids, sb.String())
					i = j
					break
				} else {
					sb.WriteRune(msgRune[j])
				}
			}
			tag = 0
		} else {
			tag = 0
		}
	}
	return uids
}
