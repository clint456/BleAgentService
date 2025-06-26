package driver

import (
	"device-ble/internal/interfaces"
	"device-ble/pkg/ble"
	"device-ble/pkg/dataparse"
	"fmt"
	"strings"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
)

const (
	TopicBLEUp        = "edgex/service/data/device_ble/up"
	TopicBLEDown      = "edgex/service/data/device_ble/dwon"
	TopicAllStatusReq = "edgex/core/commandquery/request/all"
	TopicResponse     = "edgex/response/#"
	TopicReadingReq   = "edgex/core/command/request"
)

// CommandService 负责上行命令分发和业务处理。
type CommandService struct {
	Logger           logger.LoggingClient
	MessageBusClient interfaces.MessageBusClient
	BleController    interfaces.BLEController
}

// HandleCommand 处理命令分发。
func (cs *CommandService) HandleCommand(cmd string) {
	if strings.Contains(cmd, "allstatus") {
		// cmd数据格式：+COMMAND:allstatus
		cs.Logger.Infof("【运维 — allstatus】开始查询所有设备状态")
		data, err := cs.requestAndParseAll(TopicAllStatusReq, TopicResponse, 5*time.Second)
		if err != nil {
			cs.Logger.Errorf("【运维 — allstatus 请求&解析失败: %v", err)
			return
		}
		err = ble.SendJSONOverBLE(cs.BleController.GetQueue(), data)
		data = nil
		if err != nil {
			cs.Logger.Errorf("【运维 — allstatus】发送响应失败: %v", err)
			return
		}

	} else if strings.Contains(cmd, "monitor") {
		// cmd数据格式：monitor,<deviceNamce>,<resourceName>
		// +COMMAND:monitor,Random-Integer-Device,Int8
		cs.Logger.Infof("【运维 — monitor】开始设备监控 %v", cmd)
		parts := strings.Split(cmd, ",")
		if len(parts) >= 3 {
			deviceName := parts[1]
			resourceName := parts[2]
			// 使用 deviceName 和 resourceName 继续处理逻辑
			cs.Logger.Infof("【运维 — monitor】 device: %s, resource: %s\n", deviceName, resourceName)
			cs.Logger.Infof("【运维 — monitor】开始监控指定设备数据")
			TopicReadingRequest := fmt.Sprintf("%s/%s/%s/%s", TopicReadingReq, deviceName, resourceName, "get")
			data, err := cs.requestAndParseReading(TopicReadingRequest, TopicResponse, 5*time.Second)
			if err != nil {
				cs.Logger.Errorf("【运维 —  monitor】 请求&解析失败: %v", err)
				return
			}
			err = ble.SendJSONOverBLE(cs.BleController.GetQueue(), data)
			data = nil
			if err != nil {
				cs.Logger.Errorf("【运维 —  monitor】发送响应失败: %v", err)
				return
			}
		} else {
			cs.Logger.Errorf("【运维 — monitor】命令格式错误，应为：monitor,<deviceName>,<resourceName>")
			return
		}
	} else {
		cs.Logger.Warnf("命名不支持！！")
		err := ble.SendJSONOverBLE(cs.BleController.GetQueue(), "命名不支持！！")
		if err != nil {
			cs.Logger.Errorf("【运维——status】发送响应失败: %v", err)
			return
		}
		return
	}
}

func (cs *CommandService) requestAndParseAll(
	reqTopic string,
	respTopic string,
	timeout time.Duration,
) (interface{}, error) {

	cs.Logger.Infof("【运维】准备订阅响应主题: %s", respTopic)
	if err := cs.MessageBusClient.SubscribeResponse(respTopic); err != nil {
		cs.Logger.Errorf("订阅响应失败: %v", err)
		return nil, err
	}
	cs.Logger.Infof("【运维】订阅响应成功: %s", respTopic)

	cs.MessageBusClient.SetTimeout(timeout)

	resp, err := cs.MessageBusClient.Request(reqTopic, "")
	if err != nil {
		cs.Logger.Errorf("请求失败 [%s]: %v", reqTopic, err)
		return nil, err
	}
	cs.Logger.Infof("【运维】收到响应")

	data, err := dataparse.ParseDeviceLists(&resp)
	if err != nil {
		cs.Logger.Errorf("响应数据解析失败: %v", err)
		return nil, err
	}
	return data, nil
}

func (cs *CommandService) requestAndParseReading(
	reqTopic string,
	respTopic string,
	timeout time.Duration,
) (interface{}, error) {

	cs.Logger.Infof("【运维】准备订阅响应主题: %s", respTopic)
	if err := cs.MessageBusClient.SubscribeResponse(respTopic); err != nil {
		cs.Logger.Errorf("订阅响应失败: %v", err)
		return nil, err
	}
	cs.Logger.Infof("【运维】订阅响应成功: %s", respTopic)

	cs.MessageBusClient.SetTimeout(timeout)

	resp, err := cs.MessageBusClient.Request(reqTopic, "")
	if err != nil {
		cs.Logger.Errorf("请求失败 [%s]: %v", reqTopic, err)
		return nil, err
	}
	cs.Logger.Infof("【运维】收到响应 %v", resp)

	data, err := dataparse.ParseReading(&resp)
	if err != nil {
		cs.Logger.Errorf("响应数据解析失败: %v", err)
		return nil, err
	}
	return data, nil
}

// AgentService 负责透明代理上行数据处理。
type AgentService struct {
	Logger           logger.LoggingClient
	MessageBusClient interfaces.MessageBusClient
}

// HandleAgentData 处理透明代理数据。
func (as *AgentService) HandleAgentData(data string) {
	if data == "" {
		return
	}
	type Payload struct {
		Timestamp int64
		Data      string
	}
	as.Logger.Infof("【透明代理（↑）】：收到上行数据: %s", data)
	p := Payload{
		Timestamp: time.Now().UnixNano(),
		Data:      data,
	}
	if as.MessageBusClient != nil {
		err := as.MessageBusClient.Publish(TopicBLEUp, p)
		if err != nil {
			as.Logger.Errorf("【透明代理（↑）】转发至消息总线失败 ❌: %v", err)
		} else {
			as.Logger.Infof("【透明代理（↑）】转发至消息总线成功 ✔")
		}
	}
}
