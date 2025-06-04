package driver

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// --------------------------
// 类型定义
// --------------------------

type BLECommand struct {
	Type      string      `json:"type"`            // "command"
	Target    string      `json:"target"`          // EdgeX Device 名称
	Operation string      `json:"operation"`       // "read" 或 "write"
	Resource  string      `json:"resource"`        // Resource 名
	Value     interface{} `json:"value,omitempty"` // 写操作时需要
}

type BLEResponse struct {
	Status    string      `json:"status"` // OK / ERROR
	Target    string      `json:"target"`
	Resource  string      `json:"resource"`
	Value     interface{} `json:"value,omitempty"` // 读到的值
	Error     string      `json:"error,omitempty"` // 错误信息
	Timestamp int64       `json:"timestamp"`       // 时间戳（毫秒）
}

// --------------------------
// 核心配置
// --------------------------

const (
	coreCommandHost = "http://localhost:59882"
	timeout         = 3 * time.Second
)

// --------------------------
// 消息主入口
// --------------------------

func HandleBLEMessage(msg string) (string, error) {
	var cmd BLECommand
	if err := json.Unmarshal([]byte(msg), &cmd); err != nil {
		return marshalResponse(buildError(cmd, fmt.Errorf("JSON解析失败: %w", err))), nil
	}

	switch cmd.Operation {
	case "read":
		return handleRead(cmd)
	case "write":
		return handleWrite(cmd)
	default:
		return marshalResponse(buildError(cmd, errors.New("不支持的操作类型"))), nil
	}
}

// --------------------------
// 处理读取操作
// --------------------------

func handleRead(cmd BLECommand) (string, error) {
	url := fmt.Sprintf("%s/api/v3/device/name/%s/command/%s", coreCommandHost, cmd.Target, cmd.Resource)

	client := http.Client{Timeout: timeout}
	resp, err := client.Get(url)
	if err != nil {
		return marshalResponse(buildError(cmd, err)), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return marshalResponse(buildError(cmd, fmt.Errorf("请求失败: %s", string(body)))), nil
	}

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return marshalResponse(buildError(cmd, fmt.Errorf("解析响应失败: %w", err))), nil
	}

	val := extractValue(data)
	return marshalResponse(BLEResponse{
		Status:    "OK",
		Target:    cmd.Target,
		Resource:  cmd.Resource,
		Value:     val,
		Timestamp: time.Now().UnixMilli(),
	}), nil
}

// --------------------------
// 处理写入操作
// --------------------------

func handleWrite(cmd BLECommand) (string, error) {
	url := fmt.Sprintf("%s/api/v3/device/name/%s/command/%s", coreCommandHost, cmd.Target, cmd.Resource)

	payload := map[string]interface{}{
		"value": cmd.Value,
	}
	bodyBytes, _ := json.Marshal(payload)

	client := http.Client{Timeout: timeout}
	req, _ := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return marshalResponse(buildError(cmd, err)), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return marshalResponse(buildError(cmd, fmt.Errorf("写入失败: %s", string(body)))), nil
	}

	return marshalResponse(BLEResponse{
		Status:    "OK",
		Target:    cmd.Target,
		Resource:  cmd.Resource,
		Value:     cmd.Value,
		Timestamp: time.Now().UnixMilli(),
	}), nil
}

// --------------------------
// 辅助方法
// --------------------------

func extractValue(data map[string]interface{}) interface{} {
	readings, ok := data["readings"].([]interface{})
	if !ok || len(readings) == 0 {
		return nil
	}
	first, ok := readings[0].(map[string]interface{})
	if !ok {
		return nil
	}
	return first["value"]
}

func buildError(cmd BLECommand, err error) BLEResponse {
	return BLEResponse{
		Status:    "ERROR",
		Target:    cmd.Target,
		Resource:  cmd.Resource,
		Error:     err.Error(),
		Timestamp: time.Now().UnixMilli(),
	}
}

func marshalResponse(resp BLEResponse) string {
	jsonBytes, _ := json.Marshal(resp)
	return string(jsonBytes)
}
