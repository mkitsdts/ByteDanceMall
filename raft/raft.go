package raft

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"gorm.io/gorm"
)

func InitRaftNode() *RaftNode {
	// 读取配置文件
	file, err := os.Open("configs.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var node RaftNode
	err = decoder.Decode(&node)
	if err != nil {
		panic(err)
	}
	node.state = FOLLOWER
	node.Term = TERM_MIN
	node.VoteCount = 0
	node.LeaderIP = ""
	node.Log = make([]LogEntry, 0)
	node.httpClient = http.Client{
		Timeout: 5 * time.Second,
	}
	// 监听选举请求
	http.HandleFunc("/election", node.handleElection)
	// 监听日志请求
	http.HandleFunc("/append", node.handleAppendEntries)

	return &node
}

// 持久化日志数据
func (rn *RaftNode) PersistLog(db *gorm.DB) {
	go func() {
		if rn.CommitIndex == len(rn.Log) {
			return
		}
		// 执行 sql 语句
		rn.mux.Lock()
		defer rn.mux.Unlock()
		for i := rn.CommitIndex; i < len(rn.Log); i++ {
			db.Exec(rn.Log[i].Command)
		}
	}()
}

func (rn *RaftNode) Start(db *gorm.DB) {
	go func() {
		for {
			switch rn.state {
			case FOLLOWER:
				rn.startElectionTimer()
			case CANDIDATE:
				rn.startElectionTimer()
			case LEADER:
				rn.startElectionTimer()
			}
		}
	}()
}
