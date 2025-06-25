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
	"device-ble/pkg/ble"
	"device-ble/pkg/config"
	"device-ble/pkg/mqttbus"
	"device-ble/pkg/uart"
	"fmt"
	"log"

	"github.com/edgexfoundry/device-sdk-go/v4/pkg/startup"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
)

const (
	serviceName string = "device-ble"
)

func main() {
	// 1. 初始化日志
	lc := logger.NewClient(serviceName, "INFO")

	// 2. 加载配置
	cfg, err := config.Load("./res/configBle.yaml")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("自定义的串口配置:", cfg.Serial)
	fmt.Println("自定义的MQTT客户端配置:", cfg.MQTT)

	// 3. 初始化串口
	serialPort, err := uart.NewSerialPort(cfg.Serial, lc)
	if err != nil {
		log.Fatal("创建串口实例失败:", err)
	}

	// 4. 初始化Driver（先new空Driver用于闭包回调）
	d := &driverpkg.Driver{
		Config: cfg,
	}

	// 5. 初始化串口队列，注册Driver回调
	serialQueue := uart.NewSerialQueue(
		serialPort,
		lc,
		func(cmd string) { d.HandleUpCommandCallback(cmd) },
		func(data string) { d.HandleUpAgentCallback(data) },
	)

	// 6. 初始化BLE控制器
	bleController := ble.NewBLEController(serialPort, serialQueue, lc)

	// 7. 初始化消息总线
	mqttClient, err := mqttbus.NewEdgexMessageBusClient(cfg.MQTT, lc)
	if err != nil {
		log.Fatal("MessageBusClient 创建失败:", err)
	}

	// 8. 初始化业务服务
	commandService := &driverpkg.CommandService{
		Logger:           lc,
		MessageBusClient: mqttClient,
		BleController:    bleController,
	}
	agentService := &driverpkg.AgentService{
		Logger:           lc,
		MessageBusClient: mqttClient,
	}

	// 9. 注入依赖到Driver
	d.BleController = bleController
	d.MessageBusClient = mqttClient
	d.CommandService = commandService
	d.AgentService = agentService

	// 10. 启动服务
	startup.Bootstrap(serviceName, device.Version, d)
}
