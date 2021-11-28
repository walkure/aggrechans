package common

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/slack-go/slack"
)

type UserInfo struct {
	name  map[string]*UserProfile
	api   *slack.Client
	mu    sync.Mutex
	redis *redis.Client
}

type UserProfile struct {
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
	Bot    bool   `json:"bot"`
	App    bool   `json:"app"`
}

func CreateUserInfo(api *slack.Client, redis *redis.Client) *UserInfo {
	info := UserInfo{}
	info.name = make(map[string]*UserProfile)
	info.api = api
	info.redis = redis

	return &info
}

func InitUserInfo(ctx context.Context, api *slack.Client) (*UserInfo, error) {
	info := CreateUserInfo(api, nil)

	users, err := api.GetUsersContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("err at users.list:%w", err)
	}
	for _, user := range users {
		info.setUserInfo(ctx, &user)
	}
	//api.Debugf("loaded %d users\n", len(info.name))
	fmt.Printf("loaded %d users\n", len(info.name))

	return info, nil
}

func (prof *UserProfile) IsBots() bool {
	return prof.Bot || prof.App
}

func (info *UserInfo) GetUserProfile(ctx context.Context, uid string) (*UserProfile, error) {
	if prof, ok := info.lookupUserInfo(ctx, uid); ok {
		return prof, nil
	}

	if strings.HasPrefix(uid, "B") {
		bot, err := info.api.GetBotInfoContext(ctx, uid)
		if err != nil {
			return nil, fmt.Errorf("err at bot.info(uid=%s):%w", uid, err)
		}

		return info.setBotInfo(ctx, bot), nil
	}

	user, err := info.api.GetUserInfoContext(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("err at users.info(uid=%s):%w", uid, err)
	}

	return info.setUserInfo(ctx, user), nil
}

func (info *UserInfo) HandleUserChangeEvent(ctx context.Context, ev *slack.UserChangeEvent) {
	info.setUserInfo(ctx, &ev.User)
}

func (info *UserInfo) lookupUserInfo(ctx context.Context, uid string) (*UserProfile, bool) {
	var prof *UserProfile
	ok := false

	func() {
		info.mu.Lock()
		defer info.mu.Unlock()
		prof, ok = info.name[uid]
	}()

	if ok || info.redis == nil {
		return prof, ok
	}

	result, err := info.redis.Get(ctx, uid).Result()
	if err != nil {
		return nil, false
	}

	prof = &UserProfile{}

	if err := json.Unmarshal([]byte(result), prof); err != nil {
		fmt.Fprintf(os.Stderr, "unmarshal error:%v\n", err)
		return nil, false
	}

	func() {
		info.mu.Lock()
		defer info.mu.Unlock()
		info.name[uid] = prof
	}()

	return prof, true
}

func (up UserProfile) MarshalBinary() ([]byte, error) {
	return json.Marshal(up)
}

func (info *UserInfo) setUserInfo(ctx context.Context, user *slack.User) *UserProfile {
	prof := &UserProfile{
		Name:   user.Name,
		Avatar: user.Profile.Image72,
		Bot:    user.IsBot || user.ID == "USLACKBOT",
		App:    user.IsAppUser,
	}

	func() {
		info.mu.Lock()
		defer info.mu.Unlock()
		info.name[user.ID] = prof
	}()

	if info.redis != nil {
		// discard error
		err := info.redis.Set(ctx, user.ID, prof, 0).Err()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Redis Set Error:%v\n", err)
		}
	}

	return prof
}

func (info *UserInfo) setBotInfo(ctx context.Context, bot *slack.Bot) *UserProfile {
	prof := &UserProfile{
		Name:   bot.Name,
		Avatar: bot.Icons.Image72,
		Bot:    true,
		App:    true,
	}

	func() {
		info.mu.Lock()
		defer info.mu.Unlock()
		info.name[bot.ID] = prof
	}()

	if info.redis != nil {
		// discard error
		err := info.redis.Set(ctx, bot.ID, prof, 0).Err()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Redis Set Error:%v\n", err)
		}
	}

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
