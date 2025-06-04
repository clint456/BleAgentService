// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018-2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"encoding/json"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// onIncomingDataReceived 处理通过 MQTT 接收到的消息
func (s *Driver) onIncomingDataReceived(client mqtt.Client, message mqtt.Message) {

	// 获取接收到的消息主题
	incomingTopic := message.Topic()
	// 从消息主题中移除订阅主题部分，提取元数据
	incomingTopic = strings.Replace(incomingTopic, "edgex", "", -1)
	// 获取服务配置中的订阅主题，并移除通配符 "#"
	subscribedTopic := s.serviceConfig.MQTTBrokerInfo.IncomingTopic
	subscribedTopic = strings.Replace(subscribedTopic, "#", "", -1)

	// 解析消息的 payload（JSON 格式）
	asyncData := make(map[string]interface{})
	err := json.Unmarshal(message.Payload(), &asyncData)
	if err != nil {
		s.lc.Errorf("❗️[EdgeX %v 服务数据监听] 反序列化payload失败 : %v", subscribedTopic, err)

	}

	// 记录接收到的消息信息
	s.lc.Debugf("💬[EdgeX %v 服务数据监听] topic=%v, msg=%v", subscribedTopic, message.Topic(), string(message.Payload()))
	// 创建MessageClient 并转发接收到的数据 到MessageBus 自定义主题"edgex/data/subscribedTopic"以供其它设备使用
	// 转发到 MessageBus
	err = s.publishToMessageBus(asyncData, "edgex/data"+incomingTopic)
	if err != nil {
		s.lc.Errorf("❗️[EdgeX %v 服务数据监听] 转发到 MessageBus 失败: %v", incomingTopic, err)
		return
	}
	// 将接收到的数据向蓝牙发送器异步传输数据
	s.sendToBluetoothTransmitter(asyncData)
}
