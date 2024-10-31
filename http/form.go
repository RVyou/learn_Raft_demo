package http

type RequestVote struct {
	Term        int    `json:"term"`
	CandidateId string `json:"cid"`

	LastLogIndex int `json:"last_log_index"`
	LastLogTerm  int `json:"last_log_term"`
}

type ResponseVote struct {
	Term int  `json:"term"`
	Vote bool `json:"vote"`
}

type RequestHeartbeat struct {
	Term        int    `json:"term"`
	CandidateId string `json:"cid"`

	//确认唯一日志
	PrevLogIndex int    `json:"prev_log_index"`
	PreVlogTerm  int    `json:"pre_vlog_term"`
	Entries      string `json:"entries"`

	//应用 commit index
	LeaderCommit int `json:"leader_id"`
}

type ResponseHeartbeat struct {
	Term    int  `json:"term"`
	Success bool `json:"vote"`

	ConflictIndex int `json:"conflict_index"` //附带当前日志信息
	ConflictTerm  int `json:"conflict_term"`  //附带当前日志信息
}
