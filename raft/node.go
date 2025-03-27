package raft

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"sync"
	"time"
)

type RaftNode struct {
	mux           sync.Mutex
	state         uint8
	Term          int    // 当前任期
	VoteCount     uint16 // 获得的选票数
	CommitIndex   int    // 已知的最大的已经被提交的日志条目的索引值
	LastApplied   int    // 最后被应用到状态机的日志条目索引值
	electionTimer *time.Timer
	isReceived    bool     // 是否收到信号
	LocalIP       string   `json:"local_ip"`
	NodeIP        []string `json:"node_ip"`
	LeaderIP      string   `json:"leader_ip"`
	Log           []LogEntry
}

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
	// 监听选举请求
	http.HandleFunc("/election", node.handleElection)
	// 监听日志请求
	http.HandleFunc("/append", node.handleAppendEntries)
	// 监听心跳
	http.HandleFunc("/heartbeat", node.handleHeartbeat)
	// 监听提交请求
	http.HandleFunc("/commit", node.handleCommit)
	return &node
}

// 接受心跳
func (rn *RaftNode) listenSignal() {
	rn.mux.Lock()
	defer rn.mux.Unlock()

	// 随机化超时时间（150ms ~ 300ms）
	timeout := RandomDuration()

	// 停止旧的计时器
	if rn.electionTimer != nil {
		rn.electionTimer.Stop()
	}

	// 创建新计时器
	rn.electionTimer = time.AfterFunc(timeout, func() {
		if rn.isReceived {
			rn.isReceived = false
			return
		} else {
			rn.state = CANDIDATE
			return
		}
	})
}

// 开始选举
func (rn *RaftNode) startElection(httpClient *http.Client) {
	rn.mux.Lock()
	defer rn.mux.Unlock()

	rn.Term++
	rn.VoteCount = 1
	rn.LeaderIP = ""

	for _, node := range rn.NodeIP {
		// 发送选举请求
		go func() {
			// 生成选举请求
			reqInfo := ElectionReq{
				Term:         rn.Term,
				CandidateIP:  rn.LocalIP,
				LastLogIndex: len(rn.Log) - 1,
			}
			reqBody, _ := json.Marshal(reqInfo)
			req, err := http.NewRequest("POST", "http://"+node+":"+DEFAULTPORT+"/election", bytes.NewBuffer(reqBody))
			if err != nil {
				return
			}
			req.Header.Set("Content-Type", "application/json")

			// 发送选举请求
			resp, err := httpClient.Do(req)
			if err != nil {
				return
			}
			defer resp.Body.Close()
			var electionResp ElectionResp
			decoder := json.NewDecoder(resp.Body)
			err = decoder.Decode(&electionResp)
			if err != nil {
				return
			}

			// 处理选举结果
			rn.mux.Lock()
			if electionResp.Term > rn.Term {
				rn.Term = electionResp.Term
				rn.state = FOLLOWER
				rn.VoteCount = 0
			} else if electionResp.VoteGranted {
				rn.VoteCount++
				if rn.VoteCount > uint16(len(rn.NodeIP)/2) {
					rn.state = LEADER
					rn.LeaderIP = rn.LocalIP
				}
			}
			rn.mux.Unlock()
		}()
	}
}

// 发送信号
func (rn *RaftNode) sendSignal(httpClient *http.Client) {
	for _, node := range rn.NodeIP {
		size := len(rn.Log)
		go func() {
			time.Sleep(RandomDuration())
			// 生成包
			if len(rn.Log) <= size {
				return
			}

			reqInfo := AppendEntriesReq{
				Term:         rn.Term,
				LeaderIP:     rn.LocalIP,
				PrevLogIndex: len(rn.Log) - 1,
				PrevLogTerm:  rn.Log[len(rn.Log)-1].Term,
				Entries:      make([]LogEntry, 0),
				LeaderCommit: 0,
			}
			reqBody, _ := json.Marshal(reqInfo)
			// 发送心跳
			req, err := http.NewRequest("POST", "http://"+node+":"+DEFAULTPORT+"/heartbeat", bytes.NewBuffer(reqBody))
			if err != nil {
				return
			}
			resp, err := httpClient.Do(req)
			if err != nil {
				return
			}
			defer resp.Body.Close()
		}()
	}
}
