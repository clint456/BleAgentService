package dataparse

import (
	"encoding/json"
	"fmt"

	"github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"
)

// ResourceInfo 定义输出的 JSON 结构
type ResourceInfo struct {
	DeviceName  string        `json:"deviceName"`
	ProfileName string        `json:"profileName"`
	Commands    []CommandInfo `json:"commands"`
}

type CommandInfo struct {
	Name          string   `json:"name"`
	ResourceNames []string `json:"resourceNames"`
}

// ExtractProfileAndResources 解析 payload 并返回 JSON 格式的数据
func ExtractProfileAndResources(envelope *types.MessageEnvelope) ([]byte, error) {
	// 断言 Payload 为 map[string]interface{}
	payload, ok := envelope.Payload.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("payload is not a map")
	}

	// 提取 deviceCoreCommands
	deviceCoreCommands, ok := payload["deviceCoreCommands"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("deviceCoreCommands is not a slice")
	}

	// 存储结果
	var result []ResourceInfo

	// 遍历每个设备
	for _, device := range deviceCoreCommands {
		deviceMap, ok := device.(map[string]interface{})
		if !ok {
			continue
		}

		// 提取 deviceName 和 profileName
		deviceName, _ := deviceMap["deviceName"].(string)
		profileName, _ := deviceMap["profileName"].(string)

		// 创建 ResourceInfo
		resourceInfo := ResourceInfo{
			DeviceName:  deviceName,
			ProfileName: profileName,
			Commands:    []CommandInfo{},
		}

		// 提取 coreCommands
		coreCommands, ok := deviceMap["coreCommands"].([]interface{})
		if !ok {
			continue
		}

		// 遍历每个命令
		for _, cmd := range coreCommands {
			cmdMap, ok := cmd.(map[string]interface{})
			if !ok {
				continue
			}

			// 提取命令名称
			name, _ := cmdMap["name"].(string)

			// 创建 CommandInfo
			commandInfo := CommandInfo{
				Name:          name,
				ResourceNames: []string{},
			}

			// 提取 parameters
			params, ok := cmdMap["parameters"].([]interface{})
			if !ok {
				continue
			}

			// 遍历每个参数
			for _, param := range params {
				paramMap, ok := param.(map[string]interface{})
				if !ok {
					continue
				}

				// 提取 resourceName
				if resourceName, ok := paramMap["resourceName"].(string); ok {
					commandInfo.ResourceNames = append(commandInfo.ResourceNames, resourceName)
				}
			}

			// 添加到 Commands
			if len(commandInfo.ResourceNames) > 0 {
				resourceInfo.Commands = append(resourceInfo.Commands, commandInfo)
			}
		}

		// 添加到结果
		if len(resourceInfo.Commands) > 0 {
			result = append(result, resourceInfo)
		}
	}

	// 转换为 JSON
	jsonData, err := json.MarshalIndent(result, "", "  ") // 格式化输出
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result to JSON: %v", err)
	}

	fmt.Println(string(jsonData))
	return jsonData, nil
}
