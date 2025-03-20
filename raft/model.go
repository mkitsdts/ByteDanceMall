package raft

import (
	"net/http"
	"sync"
	"time"
)

const (
	FOLLOWER  uint8 = 0
	CANDIDATE uint8 = 1
	LEADER    uint8 = 2

	TERM_MIN = 0

	DEFAULTPORT = "11451"
)

type LogEntry struct {
	Term    int    `json:"term"`
	Command string `json:"command"` // 一个 sql 语句
}

type RaftNode struct {
	mux           sync.Mutex
	state         uint8
	Term          int    // 当前任期
	VoteCount     uint16 // 获得的选票数
	CommitIndex   int    // 已知的最大的已经被提交的日志条目的索引值
	httpClient    http.Client
	electionTimer *time.Timer
	LocalIP       string   `json:"local_ip"`
	NodeIP        []string `json:"node_ip"`
	LeaderIP      string   `json:"leader_ip"`
	Log           []LogEntry
}

type ElectionReq struct {
	Term         int    `json:"term"`
	CandidateIP  string `json:"candidate_ip"`
	LastLogIndex int    `json:"last_log_index"`
}

type ElectionResp struct {
	Term        int  `json:"term"`
	VoteGranted bool `json:"vote_granted"`
}

type AppendEntriesReq struct {
	Term         int        `json:"term"`
	LeaderIP     string     `json:"leader_ip"`
	PrevLogIndex int        `json:"prev_log_index"`
	PrevLogTerm  int        `json:"prev_log_term"`
	Entries      []LogEntry `json:"entries"`
	LeaderCommit int        `json:"leader_commit"`
}

type AppendEntriesResp struct {
	Term    int  `json:"term"`
	Success bool `json:"success"`
}
