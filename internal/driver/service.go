package driver

import (
	"device-ble/internal/interfaces"
	"device-ble/pkg/ble"
	"device-ble/pkg/dataparse"
	"strings"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
)

const (
	TopicBLEUp        = "edgex/service/data/device_ble/up"
	TopicBLEDown      = "edgex/service/data/device_ble/dwon"
	TopicAllStatusReq = "edgex/core/commandquery/request/all"
	TopicResponseAll  = "edgex/response/#"
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
		cs.Logger.Infof("【运维——allstatus】开始查询所有设备状态")
		err := cs.MessageBusClient.SubscribeResponse(TopicResponseAll)
		if err != nil {
			cs.Logger.Errorf("订阅响应失败: %v", err)
			return
		}
		cs.Logger.Infof("【运维——allstatus】订阅响应成功")
		cs.MessageBusClient.SetTimeout(2 * time.Second)
		resp, err := cs.MessageBusClient.Request(TopicAllStatusReq, "")
		if err != nil {
			cs.Logger.Errorf("【运维——allstatus】请求失败: %v", err)
			return
		}
		cs.Logger.Infof("【运维——allstatus】 请求系统响应:")
		data, err := dataparse.ExtractProfileAndResources(&resp)
		if err != nil {
			cs.Logger.Errorf("【运维——allstatus】数据解析失败: %v", err)
			return
		}
		err = ble.SendJSONOverBLE(cs.BleController.GetQueue(), data)
		if err != nil {
			cs.Logger.Errorf("【运维——allstatus】发送响应失败: %v", err)
			return
		}
	} else if strings.Contains(cmd, "status") {
		cs.Logger.Infof("【运维——status】发起status请求: %v", cmd)
	} else {
		cs.Logger.Warnf("命名不支持！！")
		err := ble.SendJSONOverBLE(cs.BleController.GetQueue(), "命名不支持！！")
		if err != nil {
			cs.Logger.Errorf("【运维——status】发送响应失败: %v", err)
		}
	}
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
