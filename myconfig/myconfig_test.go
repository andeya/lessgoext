package myconfig

import (
	"testing"
)

type TheOneConfig struct {
	Name string
	Age  int
	Love *love
	i    int
}

type love struct {
	Desc  string
	Deep  int
	Array []string
	m     bool
}

func TestSync(t *testing.T) {
	me := &TheOneConfig{
		Name: "henry",
		Age:  30,
		Love: &love{
			Desc:  "编程",
			Deep:  5,
			Array: []string{"1", "2", "3"},
			m:     true,
		},
		i: -1,
	}
	t.Log(Sync(me, "main"))
	t.Log(me, me.Love)
	t.Log("me.Love.Array len: ", len(me.Love.Array))
}
