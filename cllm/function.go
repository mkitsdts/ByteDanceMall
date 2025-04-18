package cllm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

func (s *ModelService) Chat(context string) (string, error) {
	// 把字符串转换成io.Reader
	data := fmt.Sprintf(`{"model": "%s", 
	"messages": [{"role": "system", "content": "%s"}], 
	"temperature": 0.1, "stream": false}`, s.Name, context)
	fmt.Println("Request data:", data)
	// 发送 POST 请求
	client := s.GetHttpClient()

	resp, err := client.Post(s.Host+"/v1/chat/completions", "application/json", bytes.NewBuffer([]byte(data)))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("读取响应体错误:", err)
		return "", err
	}
	// 解析 JSON 响应
	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		fmt.Println("解析 JSON 响应错误:", err)
		return "", err
	}
	// 检查是否有错误信息
	if errMsg, ok := result["error"]; ok {
		if errMsgMap, ok := errMsg.(map[string]any); ok {
			if message, ok := errMsgMap["message"]; ok {
				fmt.Println("返回错误消息:", message)
				return "", fmt.Errorf("error: %s", message)
			}
		}
	}
	// 检查是否有响应内容
	if choices, ok := result["choices"]; ok {
		if choicesArray, ok := choices.([]any); ok && len(choicesArray) > 0 {
			if choice, ok := choicesArray[0].(map[string]any); ok {
				if message, ok := choice["message"]; ok {
					if content, ok := message.(map[string]any)["content"]; ok {
						if contentStr, ok := content.(string); ok {
							return contentStr, nil
						}
					}
				}
			}
		}
	}
	// 如果没有找到响应内容，返回一个空字符串
	return "", nil
}
