package main

import (
	"raft/raft"
	"testing"
	"time"
)

func TestRaft2(t *testing.T) {

	data2 := map[string]string{
		"http://127.0.0.1:7777": "a",
		//"http://127.0.0.1:7778": "b",
		"http://127.0.0.1:7779": "c",
	}

	r := raft.New("b", 7778, data2)
	time.Sleep(time.Second * 10)
	r.Start(2)

	select {}
}
func TestRaft(t *testing.T) {
	data1 := map[string]string{
		//"http://127.0.0.1:7777": "a",
		"http://127.0.0.1:7778": "b",
		"http://127.0.0.1:7779": "c",
	}

	r := raft.New("a", 7777, data1)
	time.Sleep(time.Second * 10)
	r.Start(1)

	select {}
}
func TestRaft3(t *testing.T) {

	data3 := map[string]string{
		"http://127.0.0.1:7777": "a",
		"http://127.0.0.1:7778": "b",
		//"http://127.0.0.1:7779": "c",
	}

	r := raft.New("c", 7779, data3)
	time.Sleep(time.Second * 10)
	r.Start(3)
	select {}
}
