package raft

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Data struct {
	Log         []LogEntry `json:"log"`
	CurrentTerm int        `json:"currentTerm"`
	VotedFor    string     `json:"votedFor"`
}

// persist 日志保存
func (rf *Raft) persist() error {
	var d = Data{
		Log:         rf.log,
		CurrentTerm: rf.currentTerm,
		VotedFor:    rf.votedFor,
	}

	b, err := json.Marshal(d)
	if err != nil {
		return err
	}
	f, err := os.Create(fmt.Sprintf("%s.log", rf.clientName))
	if err != nil {
		return err
	}
	_, err = f.Write(b)
	if err != nil {
		return err
	}
	defer f.Close()
	return nil
}

func (rf *Raft) readPersist() {
	var d Data
	f, err := os.Open(fmt.Sprintf("%s.log", rf.clientName))
	if err != nil {
		return
	}
	b, err := io.ReadAll(f)
	if err != nil {
		return
	}
	err = json.Unmarshal(b, &d)
	if err != nil {
		return
	}
	rf.log = d.Log
	rf.votedFor = d.VotedFor
	rf.currentTerm = d.CurrentTerm
}
