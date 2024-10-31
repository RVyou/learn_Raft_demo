package raft

import (
	"fmt"
	"math/rand"
	"raft/http"
	"time"
)

// changeFollower 角色变更
func (rf *Raft) changeFollower(term int) {
	Info("[change role follower]", rf.clientName, " currentTerm:", rf.currentTerm, " Term:", term)
	if term < rf.currentTerm {
		Error("[change role follower]", "term < rf.currentTerm", rf.clientName, rf.currentTerm, term)
		return
	}
	//存储日志
	shouldPersist := rf.currentTerm != term

	rf.role = Follower
	if term > rf.currentTerm {
		rf.votedFor = ""
	}
	rf.currentTerm = term

	if shouldPersist {
		if err := rf.persist(); err != nil {
			Error("[persist]", err)
		}
	}
}

// changeCandidate 角色变更
func (rf *Raft) changeCandidate() {
	Info("[change role candidate]", rf.clientName, " currentTerm:", rf.currentTerm)
	if rf.role == Leader {
		Error("[change role candidate]", "rf.role == Leader", rf.clientName, rf.currentTerm)
		return
	}

	rf.role = Candidate
	rf.votedFor = rf.clientName
	rf.currentTerm++

	if err := rf.persist(); err != nil {
		Error("[persist]", err)
	}
}

// changeCandidate 角色变更
func (rf *Raft) changeCLeader() {

	Info("[change role Leader]", rf.clientName, " currentTerm:", rf.currentTerm, " name:", rf.role)
	if rf.role != Candidate {
		Error("[change role Leader]", "role!= Candidate ", rf.clientName, rf.currentTerm, rf.role)
		return
	}

	rf.role = Leader
	for v := range rf.peers {
		rf.nextIndex[v] = len(rf.log) //先默认自己当选时候最大日志
		rf.matchIndex[v] = 0          //需要 follower 告知
	}
}

// startElection 获取选举的票数
func (rf *Raft) startElection(Term int) {
	var (
		vote = 1
	)
	rf.m.Lock()
	defer rf.m.Unlock()
	Info("[Raft startElection]")
	//如果开启选举时候的任期 小于当前的任期 说明已经有人选举成功结束任期
	if rf.currentTerm > Term || rf.role != Candidate {
		Info("[Raft startElection end]", rf.currentTerm > Term, rf.role != Leader)
		return
	}
	l := len(rf.log) - 1
	logTerm := 0
	if l > 0 {
		logTerm = rf.log[l].Term
	}
	for v, _ := range rf.peers {
		//拿票
		go func(v1 string) {
			data, _ := rf.net.Vote(v1, rf.clientName, rf.currentTerm, l, logTerm)
			rf.m.Lock()
			defer rf.m.Unlock()

			if data.Vote {
				vote++
			} else if rf.currentTerm < data.Term {
				Info("[Raft startElection end]", rf.currentTerm, data.Term)
				rf.changeFollower(data.Term)
				return
			}
			//如果开启选举时候的任期 小于当前的任期 说明已经有人选举成功结束任期
			if rf.currentTerm > Term || rf.role != Candidate {
				Info("[Raft startElection end]", rf.currentTerm > Term, rf.role != Leader)
				return
			}

			if vote > (len(rf.peers)+1)/2 {
				rf.changeCLeader()
				go rf.heartbeatAndDataAppend(Term)
				Info("[Raft startElection end]", "success")
			}
		}(v)
	}

}

func (rf *Raft) isMoreUpToDate(candidateIndex, candidateTerm int) bool {
	l := len(rf.log)
	if l == 0 {
		return false
	}

	lastIndex, lastTerm := l-1, rf.log[l-1].Term
	Info("[Raft isMoreUpToDate]", fmt.Sprintf("candidateIndex:%d,candidateTerm:%d,lastIndex:%d, lastTerm:%d ", candidateIndex, candidateTerm, lastIndex, lastTerm))
	//任期是否相等
	if lastTerm != candidateTerm {
		return lastTerm > candidateTerm
	}
	//谁的日志更大
	return candidateIndex < lastIndex
}

// resetElectionTimeout 开启选举
func (rf *Raft) resetElectionTimeout() {
	rf.electionStart = time.Now()
	rf.electionTimeout = electionMin + time.Duration(rand.Int63()%int64(electionDifference))
}

// electionTimout 选举是否timout
func (rf *Raft) electionTimout() bool {
	return time.Since(rf.electionStart) > rf.electionTimeout
}

// ticker 计时
func (rf *Raft) ticker() {
	timeOut := time.NewTimer(time.Duration(150+rand.Int63()%450) * time.Millisecond)
	for {
		select {
		//没有接受到心跳就开启选举
		case <-timeOut.C:
			rf.m.Lock()
			if rf.role != Leader && rf.electionTimout() {
				rf.changeCandidate()
				go rf.startElection(rf.currentTerm)
			}
			rf.m.Unlock()
		}
		timeOut.Reset(time.Duration(150+rand.Int63()%450) * time.Millisecond)
	}
}

// Vote 选举投票
func (rf *Raft) Vote(vote http.RequestVote) http.ResponseVote {
	rf.m.Lock()
	defer rf.m.Unlock()
	Info("[raft vote]", fmt.Sprintf("vote:%+v,%s", vote, rf.clientName))
	var result = http.ResponseVote{
		Term: rf.currentTerm,
		Vote: false,
	}
	if rf.currentTerm > vote.Term {
		Info("[raft vote]", rf.currentTerm, vote.Term)
		return result
	}

	if rf.currentTerm < vote.Term {
		rf.changeFollower(vote.Term)
	}

	//是否投过投票
	if len(rf.votedFor) == 0 {
		//比较日志
		if rf.isMoreUpToDate(vote.LastLogIndex, vote.LastLogTerm) {
			Info("[raft vote]", rf.clientName, "vote.LastLogIndex < current.LastLogIndex")
			return result
		}
		result.Vote = true
		rf.votedFor = vote.CandidateId
		rf.resetElectionTimeout()
		if err := rf.persist(); err != nil {
			Error("[persist]", err)
		}
		Info("[raft vote]", rf.clientName, " to ->", vote.CandidateId)
	} else {
		Info("[raft vote]", rf.clientName, " to ->", rf.votedFor)
	}

	return result
}
