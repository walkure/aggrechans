package common

import (
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleDispatcher(t *testing.T) {
	os.Setenv("AGGREGATE_CHANNEL_ID", "CIDFOOBAR")
	t.Cleanup(func() { os.Unsetenv("AGGREGATE_CHANNEL_ID") })

	d, err := NewDispatcher()

	assert.Nil(t, err)
	assert.Equal(t, "common.simpleDispatcher", reflect.TypeOf(d).String())
	assert.Equal(t, "CIDFOOBAR", d.Dispatch("hoge"))
	assert.Equal(t, "CIDFOOBAR", d.Dispatch("poyo"))
}

func TestMapDispatcher(t *testing.T) {
	json := `[{"prefix": "times_",
	"cid": "CIDTIMES"
},{
	"suffix": "_zatsu",
	"cid": "CIDZATSU"
},{
	"suffix": "_foobar",
	"cid": "CIDFOOBAR"
}]`
	os.Setenv("DISPATCH_CHANNEL", json)
	t.Cleanup(func() { os.Unsetenv("DISPATCH_CHANNEL") })
	d, err := NewDispatcher()

	assert.Nil(t, err)
	assert.Equal(t, "*common.mappedDispatcher", reflect.TypeOf(d).String())
	assert.Equal(t, "", d.Dispatch("hoge"))
	assert.Equal(t, "CIDTIMES", d.Dispatch("times_poyo"))
	assert.Equal(t, "CIDTIMES", d.Dispatch("times_hoge"))
	assert.Equal(t, "CIDTIMES", d.Dispatch("times_hoge_zatsu"))
	assert.Equal(t, "CIDZATSU", d.Dispatch("timez_hoge_zatsu"))
	assert.Equal(t, "", d.Dispatch("poyo"))
	assert.Equal(t, "CIDFOOBAR", d.Dispatch("poyo_foobar"))

}

func TestNewMapDispatcher(t *testing.T) {
	json := `[{"prefix": "times_",
	"cid": "CIDTIMES"
},{
	"suffix": "_zatsu",
	"cid": "CIDZATSU"
},{
	"suffix": "_foobar",
	"cid": "CIDFOOBAR"
}]`
	os.Setenv("DISPATCH_CHANNEL", json)
	t.Cleanup(func() { os.Unsetenv("DISPATCH_CHANNEL") })

	v, err := newMapDispatcher()
	assert.Nil(t, err)
	assert.Equal(t, []dispatchRule{{pattern: "times_", cid: "CIDTIMES"}}, v.prefixRules)
	assert.Equal(t, []dispatchRule{
		{pattern: "_zatsu", cid: "CIDZATSU"},
		{pattern: "_foobar", cid: "CIDFOOBAR"},
	}, v.suffixRules)

}

func TestMapDispatcherOverlapPrefersFirstRule(t *testing.T) {
	json := `[{"prefix": "times_",
	"cid": "CIDTIMES"
},{
	"prefix": "times_foo_",
	"cid": "CIDTIMESFOO"
}]`
	os.Setenv("DISPATCH_CHANNEL", json)
	t.Cleanup(func() { os.Unsetenv("DISPATCH_CHANNEL") })

	d, err := NewDispatcher()
	assert.Nil(t, err)
	// "times_foo_bar" matches both rules; registration order must decide the winner deterministically.
	assert.Equal(t, "CIDTIMES", d.Dispatch("times_foo_bar"))
}
