package raft

import (
	"encoding/json"
	"fmt"
	"raft/http"
	"sort"
	"time"
)

// LogEntry 日志格式
type LogEntry struct {
	CommandValid bool        `json:"command_valid"` //日志是否发送合并，日志压缩等等相关自身操作
	Command      interface{} `json:"command"`       //相关操作数据
	Term         int         `json:"term"`          //日志计数
}

// Heartbeat  心跳处理
func (rf *Raft) Heartbeat(data http.RequestHeartbeat) (result http.ResponseHeartbeat) {
	rf.m.Lock()
	defer rf.resetElectionTimeout()
	defer rf.m.Unlock()

	var (
		log = make([]LogEntry, 0, 8)
	)
	result.Term = rf.currentTerm
	result.ConflictIndex = -1

	Info("[raft Heartbeat]", data.CandidateId, " to ->", rf.clientName)
	if data.Term < rf.currentTerm {
		return
	}

	if data.Term >= rf.currentTerm && rf.role != Follower {
		rf.changeFollower(data.Term)
	}
	//log 同步
	if data.PrevLogIndex > len(rf.log) {
		Error("[raft Heartbeat] data.PrevLogIndex > len(rf.log)", data.CandidateId, " to ->", rf.clientName, "len ", data.PrevLogIndex, len(rf.log))
		if len(rf.log) > 0 {
			result.ConflictIndex = len(rf.log)
			result.ConflictTerm = rf.log[len(rf.log)-1].Term
		}
		return
	}
	//是否与自己日志一致
	if len(rf.log)+1 <= data.PrevLogIndex && rf.log[data.PrevLogIndex].Term != data.PreVlogTerm {
		Error("[raft Heartbeat]", data.CandidateId, " to ->", rf.clientName, "上一个日志 term 不一致 ", rf.log[data.PrevLogIndex].Term, data.PreVlogTerm)
		result.ConflictTerm = rf.log[data.PrevLogIndex].Term
		result.ConflictIndex = rf.fistIndexFor(result.ConflictTerm)
		return
	}

	if len(data.Entries) > 0 {
		if err := json.Unmarshal([]byte(data.Entries), &log); err != nil {
			Error("[raft Heartbeat] json err：", err)
			return
		}
		rf.log = append(rf.log[:data.PrevLogIndex], log...)
	}

	result.Success = true
	//确认leader commit
	if data.LeaderCommit > rf.commitIndex || rf.commitIndex == 0 && len(data.Entries) > 0 {
		Info("[raft Heartbeat] ", data.CandidateId, " to ->", rf.clientName, "update commit index ", rf.commitIndex, data.LeaderCommit)
		rf.commitIndex = data.LeaderCommit
		rf.applyCond.Signal()
	}
	return
}

// heartbeatAndDataAppendHandle 心跳发送并且发送数据
func (rf *Raft) heartbeatAndDataAppendHandle(term int) bool {
	rf.m.Lock()
	defer rf.m.Unlock()

	if rf.role != Leader || rf.currentTerm > term {
		Info("[raft heartbeat and data append]", rf.role, rf.currentTerm, term)
		return false
	}

	rf.matchIndex[rf.clientName] = len(rf.log) - 1
	rf.nextIndex[rf.clientName] = len(rf.log)
	for v, _ := range rf.peers {
		go func(v string) {
			rf.m.Lock()
			defer rf.m.Unlock()
			log := []LogEntry{}
			prevIndex := rf.nextIndex[v]
			prevTerm := 0
			data := []byte{}
			if len(rf.log) > prevIndex {
				prevTerm = rf.log[prevIndex].Term
				data, _ = json.Marshal(append(log, rf.log[prevIndex:]...))
			} else {
				prevIndex = -1
			}

			result, _ := rf.net.Heartbeat(v, rf.clientName, term, prevIndex, prevTerm, string(data), rf.commitIndex)
			if rf.currentTerm < result.Term {
				rf.changeFollower(result.Term)
				return
			}
			//日志同步失败 需要调整  prevIndex
			if !result.Success {
				//判断日志差
				if result.ConflictIndex == -1 {
					rf.nextIndex[v] = 0
				} else {
					firstIndex := rf.fistIndexFor(result.ConflictTerm)
					if firstIndex != 0 {
						rf.nextIndex[v] = firstIndex
					} else {
						rf.nextIndex[v] = result.ConflictIndex
					}
				}
				//不能超过主日志大小
				if rf.nextIndex[v] >= prevIndex && prevIndex != -1 {
					rf.nextIndex[v] = prevIndex
				}
				Info("[raft heartbeat and data append]", fmt.Sprintf("result:%+v index:%d", result, rf.nextIndex[v]))
				return
			}
			//记录同步位置
			rf.matchIndex[v] = prevIndex + len(log)
			rf.nextIndex[v] = prevIndex + len(log) + 1
			//已经得到大部分收到信息进行确认
			majorityMatched := rf.getMajorityIndexLocked()
			if majorityMatched > rf.commitIndex {
				rf.commitIndex = majorityMatched
				rf.applyCond.Signal() //弹出一个在等待的信号
			}

		}(v)
	}

	return true
}

// 消息最大确认判断
func (rf *Raft) getMajorityIndexLocked() int {
	TempIndex := make([]int, 0, len(rf.matchIndex))
	for _, v := range rf.matchIndex {
		TempIndex = append(TempIndex, v)
	}
	sort.Ints(TempIndex)
	return TempIndex[len(rf.peers)/2]
}

// heartbeat 发送心跳
func (rf *Raft) heartbeatAndDataAppend(term int) {
	t := time.NewTicker(heartbeat)
	for {
		select {
		case <-t.C:
			if !rf.heartbeatAndDataAppendHandle(term) {
				return
			}
			t.Reset(heartbeat)
		}
	}
}

// 计算任期
func (rf *Raft) fistIndexFor(term int) int {
	for i, v := range rf.log {
		if v.Term == term {
			return i
		} else if v.Term > term {
			break
		}
	}
	return 0
}
