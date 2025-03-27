package raft

const (
	FOLLOWER  uint8 = 0
	CANDIDATE uint8 = 1
	LEADER    uint8 = 2

	TERM_MIN = 0

	DEFAULTPORT = "11451"

	LOG_PATH = "log.json"
)

const (
	MIN_DURATION = 100
	MAX_DURATION = 200
)

type Command struct {
	OpType string `json:"op_type"`
	SQL    string `json:"sql"`
	Value  []byte `json:"value"`
}

type LogEntry struct {
	Term int     `json:"term"`
	Comm Command `json:"command"` // 日志条目
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
