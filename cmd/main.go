// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2021 Jiangxing Intelligence Ltd
//
// SPDX-License-Identifier: Apache-2.0

// This package provides device service of a uart devices.
package main

import (
	device "device-ble"
	driverpkg "device-ble/internal/driver"
	"device-ble/internal/interfaces"
	"device-ble/pkg/ble"
	"device-ble/pkg/dataparse"
	"device-ble/pkg/mqttbus"
	"device-ble/pkg/uart"
	"encoding/json"
	"time"

	"github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"

	"github.com/edgexfoundry/device-sdk-go/v4/pkg/startup"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
)

const (
	serviceName string = "device-ble"
)

func main() {
	// 1. 组装配置
	config := &driverpkg.Config{
		Serial: interfaces.SerialConfig{
			PortName:    "/dev/ttyS3",
			BaudRate:    115200,
			ReadTimeout: 100,
		},
		MQTT: interfaces.MQTTConfig{
			Host:     "localhost",
			Port:     1883,
			Protocol: "tcp",
			ClientID: "device-ble",
			QoS:      1,
			Username: "",
			Password: "",
		},
	}

	// 2. 组装 logger
	log := logger.NewClient(serviceName, "DEBUG")

	// 3. 组装串口与队列
	serialPort, err := uart.NewSerialPort(uart.SerialPortConfig{
		PortName:    config.Serial.PortName,
		BaudRate:    config.Serial.BaudRate,
		ReadTimeout: time.Duration(config.Serial.ReadTimeout),
	}, log)
	if err != nil {
		panic(err)
	}
	serialQueue := uart.NewSerialQueue(serialPort, log)

	// 4. 组装 BLE 控制器
	bleController := ble.NewBLEController(serialPort, serialQueue, log)

	// 5. 组装消息总线客户端
	mqttCfg := config.GetMQTTConfig()
	cfgMap := map[string]interface{}{
		"Host":     mqttCfg.Host,
		"Port":     mqttCfg.Port,
		"Protocol": mqttCfg.Protocol,
		"ClientID": mqttCfg.ClientID,
		"QoS":      mqttCfg.QoS,
		"Username": mqttCfg.Username,
		"Password": mqttCfg.Password,
	}
	subscribeTopics := []string{"edgex/events/#"}

	msgBusImpl, err := mqttbus.NewEdgexMessageBusClient(cfgMap, log)
	if err != nil {
		panic(err)
	}
	msgBus := msgBusImpl

	handler := func(topic string, envelope types.MessageEnvelope) error {
		var data map[string]interface{}
		if err := json.Unmarshal(envelope.Payload.([]byte), &data); err != nil {
			log.Errorf("解析消息失败: %v", err)
			return err
		}
		log.Infof("收到MQTT消息: topic=%s, payload=%s", topic, string(envelope.Payload.([]byte)))
		// 发布到 MessageBus
		if err := dataparse.PublishToMessageBus(msgBus, data, topic); err != nil {
			log.Errorf("转发到MessageBus失败: %v", err)
			return err
		}
		// 发送到 BLE
		dataparse.SendToBlE(bleController, data)
		return nil
	}

	if err := msgBus.Subscribe(subscribeTopics, handler); err != nil {
		panic(err)
	}

	// 6. 注入依赖
	d := driverpkg.Driver{
		Config:           config,
		BleController:    bleController,
		MessageBusClient: msgBus,
	}

	startup.Bootstrap(serviceName, device.Version, &d)
}
