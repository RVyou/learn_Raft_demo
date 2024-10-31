package raft

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	http2 "raft/http"
	"strconv"
)

type RPCServices struct {
}

func servicesNew(port int, ra *Raft) {
	Services := http.NewServeMux()
	Services.HandleFunc("/heartbeat", func(w http.ResponseWriter, r *http.Request) {
		var (
			request http2.RequestHeartbeat
			b, _    = io.ReadAll(r.Body)
		)
		_ = json.Unmarshal(b, &request)
		result, _ := json.Marshal(ra.Heartbeat(request))
		w.Write(result)
	})
	//
	Services.HandleFunc("/vote", func(w http.ResponseWriter, r *http.Request) {
		var (
			request http2.RequestVote
			b, _    = io.ReadAll(r.Body)
		)
		_ = json.Unmarshal(b, &request)
		result, _ := json.Marshal(ra.Vote(request))
		w.Write(result)
	})
	if err := http.ListenAndServe("127.0.0.1:"+strconv.Itoa(port), Services); err != nil {
		log.Fatal(err)
	}
}
