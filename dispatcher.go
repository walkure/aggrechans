package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

type ChannelDispatcher interface {
	Dispatch(chanName string) string
	Rules() string
}

type simpleDispatcher struct {
	chanId string
}

type dispatchRule struct {
	pattern string
	cid     string
}

type mappedDispatcher struct {
	suffixRules []dispatchRule
	prefixRules []dispatchRule
}

func (d simpleDispatcher) Dispatch(chanName string) string {
	return d.chanId
}

func (d simpleDispatcher) Rules() string {
	return fmt.Sprintf("send every message to:%s", d.chanId)
}

func (d *mappedDispatcher) Dispatch(chanName string) string {
	for _, r := range d.prefixRules {
		if strings.HasPrefix(chanName, r.pattern) {
			return r.cid
		}
	}

	for _, r := range d.suffixRules {
		if strings.HasSuffix(chanName, r.pattern) {
			return r.cid
		}
	}

	return ""
}

func (d *mappedDispatcher) Rules() string {
	var rules []string

	for _, r := range d.prefixRules {
		rules = append(rules, fmt.Sprintf("prefix[%s]->[%s]", r.pattern, r.cid))
	}

	for _, r := range d.suffixRules {
		rules = append(rules, fmt.Sprintf("suffix[%s]->[%s]", r.pattern, r.cid))
	}

	return strings.Join(rules, "\n")
}

func NewDispatcher() (ChannelDispatcher, error) {

	md, err := newMapDispatcher()
	if md != nil {
		return md, nil
	}

	if err != nil {
		fmt.Printf("Dispatcher disabled.:%+v\n", err)
	}

	aggChan := os.Getenv("AGGREGATE_CHANNEL_ID")
	if aggChan != "" {
		return simpleDispatcher{chanId: aggChan}, nil
	}

	return nil, errors.New("no dispatch info found")

}

type dispatchInfo []struct {
	ChannelId string `json:"cid"`
	Prefix    string `json:"prefix,omitempty"`
	Suffix    string `json:"suffix,omitempty"`
}

func newMapDispatcher() (*mappedDispatcher, error) {
	var result dispatchInfo
	dispatchJson := os.Getenv("DISPATCH_CHANNEL")
	if dispatchJson == "" {
		return nil, errors.New("DISPATCH_CHANNEL not found")
	}

	if err := json.Unmarshal([]byte(dispatchJson), &result); err != nil {
		if err, ok := err.(*json.SyntaxError); ok {
			return nil, fmt.Errorf("JSON Syntax error:%w", err)
		}
		return nil, fmt.Errorf("JSON unmarshal error:%w", err)
	}

	md := mappedDispatcher{}
	for _, v := range result {
		if v.ChannelId == "" {
			continue
		}

		if v.Prefix != "" {
			md.prefixRules = append(md.prefixRules, dispatchRule{pattern: v.Prefix, cid: v.ChannelId})
		}

		if v.Suffix != "" {
			md.suffixRules = append(md.suffixRules, dispatchRule{pattern: v.Suffix, cid: v.ChannelId})
		}
	}

	if len(md.prefixRules)+len(md.suffixRules) == 0 {
		return nil, errors.New("no dispatch rules found")
	}

	return &md, nil
}
