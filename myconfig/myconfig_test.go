package myconfig

import (
	"testing"
)

type TheOne struct {
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
	me := &TheOne{
		Name: "henry",
		Age:  30,
		Love: &love{
			Desc: "编程",
			Deep: 5,
			m:    true,
		},
		i: -1,
	}
	t.Log(Sync(me, "main"))
	t.Log(me, me.Love)
}
