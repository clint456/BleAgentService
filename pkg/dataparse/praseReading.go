package dataparse

import (
	"encoding/json"
	"fmt"

	"github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"
)

type ReadingSummary struct {
	ResourceName string `json:"resourceName"`
	Value        string `json:"value"`
	ValueType    string `json:"valueType"`
	SourceName   string `json:"sourceName"`
}

func ParseReading(envelope *types.MessageEnvelope) (*ReadingSummary, error) {
	// Step 1: 检查 Payload 类型并解码
	var payload map[string]interface{}
	if bytes, ok := envelope.Payload.([]byte); ok {
		// 如果 Payload 是 []byte，解码为 map
		err := json.Unmarshal(bytes, &payload)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal payload: %v", err)
		}
	} else if m, ok := envelope.Payload.(map[string]interface{}); ok {
		payload = m
	} else {
		return nil, fmt.Errorf("payload is not a map or byte slice")
	}

	// Step 2: 提取 event 数据
	event, ok := payload["event"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("event missing or wrong type")
	}

	sourceName, _ := event["sourceName"].(string)

	readings, ok := event["readings"].([]interface{})
	if !ok || len(readings) == 0 {
		return nil, fmt.Errorf("readings missing or empty")
	}

	first, ok := readings[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid reading format")
	}

	resourceName, _ := first["resourceName"].(string)
	value, _ := first["value"].(string)
	valueType, _ := first["valueType"].(string)

	// Step 3: 组装结果结构体
	result := &ReadingSummary{
		ResourceName: resourceName,
		Value:        value,
		ValueType:    valueType,
		SourceName:   sourceName,
	}

	// Step 4: 打印格式化 JSON
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return result, fmt.Errorf("failed to marshal result: %v", err)
	}

	fmt.Println("【解析读取结果】")
	fmt.Println(string(jsonData))

	return result, nil
}
