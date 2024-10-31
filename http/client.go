package http

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type RPCClient struct {
}

func (c *RPCClient) ping() {

}

func (c *RPCClient) Vote(url, name string, term int, index, indexTerm int) (ResponseVote, int) {
	var (
		clt  http.Client
		data = RequestVote{
			Term:         term,
			CandidateId:  name,
			LastLogIndex: index,
			LastLogTerm:  indexTerm,
		}
		resp ResponseVote
	)
	d, _ := json.Marshal(data)
	r, e := clt.Post(url+"/vote", "application/json", bytes.NewBuffer(d))
	if e != nil {
		return ResponseVote{}, 0
	}
	defer r.Body.Close()
	result, _ := io.ReadAll(r.Body)
	_ = json.Unmarshal(result, &resp)

	return resp, r.StatusCode
}

func (c *RPCClient) Heartbeat(url, name string, term int, prevIndex, prevTerm int, log string, leaderCommit int) (ResponseHeartbeat, int) {
	var (
		clt  http.Client
		data = RequestHeartbeat{
			Term:         term,
			CandidateId:  name,
			PreVlogTerm:  prevTerm,
			PrevLogIndex: prevIndex,
			Entries:      log,
			LeaderCommit: leaderCommit,
		}
		resp = ResponseHeartbeat{}
	)
	d, _ := json.Marshal(data)

	r, e := clt.Post(url+"/heartbeat", "application/json", bytes.NewBuffer(d))
	if e != nil {
		return resp, 0
	}
	defer r.Body.Close()
	result, _ := io.ReadAll(r.Body)
	_ = json.Unmarshal(result, &resp)
	//投票
	//比较
	return resp, r.StatusCode
}
