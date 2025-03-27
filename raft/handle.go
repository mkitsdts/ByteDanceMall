package raft

import (
	"encoding/json"
	"net/http"
)

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
		if rn.electionTimer != nil {
			rn.electionTimer.Stop()
		}
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
		rn.persistLog()
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

func (rn *RaftNode) handleCommit(w http.ResponseWriter, r *http.Request) {
	rn.mux.Lock()
	defer rn.mux.Unlock()

}

func (rn *RaftNode) handleHeartbeat(w http.ResponseWriter, r *http.Request) {

}
