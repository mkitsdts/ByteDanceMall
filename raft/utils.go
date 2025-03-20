package raft

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"net/http"
	"time"
)

const (
	MIN_DURATION = 150
	MAX_DURATION = 300
)

// 生成随机数时间间隔
func RandomDuration() time.Duration {
	return time.Duration(MIN_DURATION + rand.Intn(MAX_DURATION-MIN_DURATION))
}

// 选举倒计时
func (rn *RaftNode) startElectionTimer() {
	rn.mux.Lock()
	defer rn.mux.Unlock()

	// 随机化超时时间（150ms ~ 300ms）
	timeout := RandomDuration()

	// 停止旧的计时器
	if rn.electionTimer != nil {
		rn.electionTimer.Stop()
	}

	// 创建新计时器，超时后触发选举
	rn.electionTimer = time.AfterFunc(timeout, func() {
		rn.startElection()
	})
}

// 开始选举
func (rn *RaftNode) startElection() {
	rn.mux.Lock()

	// 状态切换为候选人
	rn.state = CANDIDATE
	rn.Term++
	rn.VoteCount = 1
	rn.LeaderIP = ""
	rn.mux.Unlock()

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
			resp, err := rn.httpClient.Do(req)
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

// 处理选举请求
func (rn *RaftNode) handleElection(w http.ResponseWriter, r *http.Request) {
	// 解析数据
	var reqInfo ElectionReq
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&reqInfo)
	if err != nil {
		return
	}

	// 处理选举请求
	rn.mux.Lock()
	if reqInfo.Term < rn.Term {
		// 拒绝投票
		respInfo := ElectionResp{
			Term:        rn.Term,
			VoteGranted: false,
		}
		respBody, _ := json.Marshal(respInfo)
		w.Write(respBody)
		rn.mux.Unlock()
		return
	} else if reqInfo.Term > rn.Term {
		// 更新自身状态
		rn.Term = reqInfo.Term
		rn.state = FOLLOWER
		rn.LeaderIP = ""
		rn.VoteCount = 0
		// 返回投票
		respInfo := ElectionResp{
			Term:        rn.Term,
			VoteGranted: true,
		}
		respBody, _ := json.Marshal(respInfo)
		w.Write(respBody)
		rn.mux.Unlock()
		return
	} else {
		if rn.state == CANDIDATE {
			if len(rn.Log)-1 > reqInfo.LastLogIndex {
				// 拒绝投票
				respInfo := ElectionResp{
					Term:        rn.Term,
					VoteGranted: false,
				}
				respBody, _ := json.Marshal(respInfo)
				w.Write(respBody)
				rn.mux.Unlock()
				return
			} else {
				// 投票
				respInfo := ElectionResp{
					Term:        rn.Term,
					VoteGranted: true,
				}
				respBody, _ := json.Marshal(respInfo)
				w.Write(respBody)
				rn.LeaderIP = reqInfo.CandidateIP
				rn.mux.Unlock()
				return
			}
		}
	}
}

// 发送心跳或数据
func (rn *RaftNode) sendHeartbeat() {
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
			resp, err := rn.httpClient.Do(req)
			if err != nil {
				return
			}
			defer resp.Body.Close()
		}()
	}
}

// 处理日志
func (rn *RaftNode) handleAppendEntries(w http.ResponseWriter, r *http.Request) {
	rn.mux.Lock()
	defer rn.mux.Unlock()
	// 解析数据
	var reqInfo AppendEntriesReq
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&reqInfo); err != nil {
		rn.mux.Unlock()
		return
	}
	// 判断是心跳还是日志
	if len(reqInfo.Entries) == 0 {
		// 检查日志一致性
		if reqInfo.PrevLogIndex < len(rn.Log) {
			// 拒绝心跳
			var respInfo AppendEntriesResp
			respInfo.Term = rn.Term
			respInfo.Success = false
			respBody, _ := json.Marshal(respInfo)
			w.Write(respBody)
			return
		}
		// 检查任期
		if reqInfo.Term < rn.Term {
			// 拒绝心跳
			var respInfo AppendEntriesResp
			respInfo.Term = rn.Term
			respInfo.Success = false
			respBody, _ := json.Marshal(respInfo)
			w.Write(respBody)
			return
		}
		// 更新自身状态
		rn.Term = reqInfo.Term
		rn.state = FOLLOWER
		rn.LeaderIP = reqInfo.LeaderIP
		if reqInfo.LeaderCommit > rn.CommitIndex {
			rn.CommitIndex = min(reqInfo.LeaderCommit, len(rn.Log)-1)
		}
		// 重置计时器

		// 返回心跳
		var respInfo AppendEntriesResp
		respInfo.Term = rn.Term
		respInfo.Success = true
		respBody, _ := json.Marshal(respInfo)
		w.Write(respBody)
	} else {
		// 日志
		// 处理日志
		if reqInfo.Term < rn.Term {
			// 拒绝日志
			var respInfo AppendEntriesResp
			respInfo.Term = rn.Term
			respInfo.Success = false
			respBody, _ := json.Marshal(respInfo)
			w.Write(respBody)
			return
		}
		// 检查日志一致性
		if rn.Log[reqInfo.PrevLogIndex].Term != reqInfo.PrevLogTerm {
			// 拒绝日志
			var respInfo AppendEntriesResp
			respInfo.Term = rn.Term
			respInfo.Success = false
			respBody, _ := json.Marshal(respInfo)
			w.Write(respBody)
			return
		}
		// 更新日志
		rn.Log = append(rn.Log, reqInfo.Entries...)
		// 更新提交索引
		if reqInfo.LeaderCommit > rn.Log[len(rn.Log)-1].Term {
			rn.Log[len(rn.Log)-1].Term = reqInfo.LeaderCommit
		}
		// 返回日志
		var respInfo AppendEntriesResp
		respInfo.Term = rn.Term
		respInfo.Success = true
		respBody, _ := json.Marshal(respInfo)
		w.Write(respBody)
	}
}
