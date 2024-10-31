package raft

import (
	"fmt"
	"raft/http"
	"sync"
	"time"
)

type role int

// Role 角色
// Follower会只会接收请求，当Follower接受不到请求时，会变成一个Candidate发起一次Election，获得绝大数票的Candidate会成为Leader。在Election的过程中会遵循下面几条原则：
//
// 只有当一个Candidate得到过半的选票时才能赢得选举，每一个节点按照先到先得的方式，最多投票给一位Candidate。
// 在等待投票结果的过程中，如果Candidate收到其他节点发送的心跳信息（AppendEntries RPC），并检查心跳信息中的任期不比自己小，则自己变为Follower，听从新上任的Leader的指挥。
// 当未获得结果时，会重新发起一次Election，Raft为了避免wait时间过长，Leader不可用的超时时长随机获得，以避免同一时间宕机，同样的Candidate的拉票也是随机发送请求。
const (
	Leader    role = iota + 1 //领导人
	Candidate                 //候选人
	Follower                  //追随者
)

var (
	electionMin        time.Duration = time.Millisecond * 200
	electionDifference time.Duration = time.Millisecond * 150
	heartbeat          time.Duration = time.Millisecond * 80
)

type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int

	// For PartD:
	SnapshotValid bool
	Snapshot      []byte
	SnapshotTerm  int
	SnapshotIndex int
}
type Raft struct {
	m   sync.Mutex
	net http.RPCClient //使用http 代替rpc

	clientName      string
	peers           map[string]string //其他名字和服务地址映射
	role            role
	currentTerm     int           //当前任期
	votedFor        string        //Leader是谁
	electionStart   time.Time     //选举开始时间
	electionTimeout time.Duration //选举超时差值

	log        []LogEntry     //本地日志
	nextIndex  map[string]int //Leader发送日志
	matchIndex map[string]int //Leader确认日志

	applyCh     chan ApplyMsg //日志应用管道
	commitIndex int
	lastApplied int
	applyCond   *sync.Cond
}

func New(name string, port int, peers map[string]string) *Raft {
	var (
		r = Raft{
			role:        Follower, //初始化状态
			votedFor:    "",
			net:         http.RPCClient{},
			clientName:  name,
			peers:       peers,
			currentTerm: 0,
			commitIndex: 0,
			log:         make([]LogEntry, 0, 100),
			nextIndex:   make(map[string]int, len(peers)),
			matchIndex:  make(map[string]int, len(peers)),
		}
		w sync.WaitGroup
	)
	Name = name
	r.applyCond = sync.NewCond(&r.m)
	r.applyCh = make(chan ApplyMsg, 1000)
	r.readPersist()
	w.Add(1)
	go func() {
		w.Done()
		servicesNew(port, &r)
	}()

	w.Wait()
	//读取以前的状态进行掩盖
	go r.ticker()
	go r.applicationTicker()
	return &r
}

func (rf *Raft) Start(Command interface{}) (int, int, bool) {
	rf.m.Lock()
	defer rf.m.Unlock()
	if rf.role != Leader {
		Info("[raft add data]", "not leader")
		return 0, 0, false
	}
	rf.log = append(rf.log, LogEntry{
		CommandValid: true,
		Command:      Command,
		Term:         rf.currentTerm,
	})
	Info("[raft add data]", fmt.Sprintf("rf.log.length:%d,currentTerm:%d", len(rf.log), rf.currentTerm))
	return len(rf.log) - 1, rf.currentTerm, true
}
