package raft

import (
	"encoding/json"
	"os"
)

// 暂时不能保存大量数据
// 保存日志到磁盘
func (rn *RaftNode) persistLog() {

	rn.mux.Lock()
	defer rn.mux.Unlock()

	if rn.CommitIndex-rn.LastApplied <= 0 {
		return
	}

	// 判断文件是否存在
	_, err := os.Stat(LOG_PATH)
	if err != nil {
		// 创建文件
		file, err := os.Create(LOG_PATH)
		if err != nil {
			return
		}
		defer file.Close()
	}
	// 打开文件
	file, err := os.OpenFile(LOG_PATH, os.O_WRONLY, 0666)
	if err != nil {
		return
	}
	defer file.Close()

	var logs []LogEntry
	// 读取日志
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&logs)
	if err != nil {
		return
	}

	// 写入日志
	for i := rn.LastApplied; i < rn.CommitIndex; i++ {
		logs = append(logs, rn.Log[i])
	}

	// 保存日志
	encoder := json.NewEncoder(file)
	err = encoder.Encode(logs)
	if err != nil {
		return
	}
}
