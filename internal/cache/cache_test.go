package cache

import (
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	type Value struct {
		Name string
	}
	c := New[string, Value]()
	c.Set("chaotingchen10@gmail.com", Value{Name: "Tim"}, time.Second*30)
	c.Set("dev@dev.com", Value{Name: "dev"}, time.Second*30)

	c.Pop("dev@dev.com")
	v, ok := c.Get("dev@dev.com")
	if ok {
		t.Error("(Error) expected value to be deleted")
	} else {
		t.Log("(Success) Poped value: ", v)
	}

	v, ok = c.Get("chaotingchen10@gmail.com")
	if !ok {
		t.Error("(Error) expected value to be found")
	} else {
		t.Log("(Success) Found: ", v)
	}

	time.Sleep(time.Second * 35)
	v, ok = c.Get("chaotingchen10@gmail.com")
	if ok {
		t.Error("(Error) expected value to be expired")
	} else {
		t.Log("(Success) Expired")
	}
}
