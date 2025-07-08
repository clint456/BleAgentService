// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018 Canonical Ltd
// Copyright (C) 2018-2019 IOTech Ltd
// Copyright (C) 2021 Jiangxing Intelligence Ltd
// Copyright (C) 2022 HCL Technologies Ltd
//
// SPDX-License-Identifier: Apache-2.0

// Package driver this package provides an UART implementation of
// ProtocolDriver interface.
//
// CONTRIBUTORS              COMPANY
//===============================================================
// 1. Sathya Durai           HCL Technologies
// 2. Sudhamani Bijivemula   HCL Technologies
// 3. Vediyappan Villali     HCL Technologies
// 4. Vijay Annamalaisamy    HCL Technologies
//
//

package driver

import (
	"device-ble/cmd/config"
	internalif "device-ble/internal/interfaces"
	"log"

	"device-ble/pkg/ble"
	"device-ble/pkg/mqttbus"
	"device-ble/pkg/uart"
	errorDefault "errors"
	"fmt"
	"time"

	"github.com/spf13/cast"

	edgexif "github.com/edgexfoundry/device-sdk-go/v4/pkg/interfaces"
	dsModels "github.com/edgexfoundry/device-sdk-go/v4/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
	"github.com/tarm/serial"
)

// Driver BLE代理服务驱动程序，协调各组件初始化和生命周期管理。
type Driver struct {
	// EdgeX SDK相关
	sdk      edgexif.DeviceServiceSDK
	logger   logger.LoggingClient
	asyncCh  chan<- *dsModels.AsyncValues
	deviceCh chan<- []dsModels.DiscoveredDevice

	// 核心组件
	BleController    internalif.BLEController
	MessageBusClient internalif.MessageBusClient

	// 业务服务
	CommandService *CommandService
	AgentService   *AgentService
}

// Initialize 初始化设备服务
/*
在 Initialize 方法执行时：
设备配置文件（如 devices.yaml）尚未加载，设备实例还未添加到 EdgeX 中，只能访问服务级别的配置，不能访问具体的设备配置
*/
func (d *Driver) Initialize(sdk edgexif.DeviceServiceSDK) error {
	d.sdk = sdk
	d.logger = sdk.LoggingClient()
	d.asyncCh = sdk.AsyncValuesChannel()
	d.deviceCh = sdk.DiscoveredDeviceChannel()
	return nil
}

// Start 启动设备服务。
func (d *Driver) Start() error {
	// 获取 UART 配置信息
	// 通过结构体字段访问 Protocols
	var deviceLocation string
	var baudRate int
	var readTimeout int
	uartConfig, err := d.sdk.GetDeviceByName("device-ble")
	if err != nil {
		d.logger.Errorf("加载服务配置失败！")
	}
	for i, protocol := range uartConfig.Protocols {
		deviceLocation = fmt.Sprintf("%v", protocol["deviceLocation"])
		baudRate, _ = cast.ToIntE(protocol["baudRate"])
		readTimeout, _ = cast.ToIntE(protocol["readTimeout"])
		d.logger.Debugf("Driver.HandleReadCommands(): protocol = %v, device location = %v, baud rate = %v readTimeout=%v", i, deviceLocation, baudRate, readTimeout)
	}

	serialPort, err := uart.NewSerialPort(serial.Config{
		Name:        deviceLocation,
		Baud:        baudRate,
		ReadTimeout: time.Duration(readTimeout) * time.Millisecond,
	}, d.logger)
	if err != nil {
		log.Fatal("创建串口实例失败:", err)
	}
	// 初始化串口队列，注册Driver回调
	serialQueue := uart.NewSerialQueue(
		serialPort,
		d.logger,
		func(cmd string) { d.HandleUpCommandCallback(cmd) },
		func(data string) { d.HandleUpAgentCallback(data) },
		5,
	)

	// 初始化BLE控制器
	bleController := ble.NewBLEController(serialPort, serialQueue, d.logger)

	// 初始化BLE设备为外围设备模式
	if err := bleController.InitializeAsPeripheral(); err != nil {
		log.Fatal("BLE设备初始化失败:", err)
	}

	// 加载自定义MQTT配置
	cfg, err := config.LoadConfig("./res/configuration.yaml")
	if err != nil {
		log.Fatal("MessageBusClient 获取自定义配置失败:", err)
	}
	d.logger.Debugf("自定义Mqtt服务配置: %v\n", cfg)
	// 初始化消息总线
	mqttClient, err := mqttbus.NewEdgexMessageBusClient(cfg, d.logger)
	if err != nil {
		log.Fatal("MessageBusClient 创建失败:", err)
	}

	// 初始化业务服务
	commandService := &CommandService{
		Logger:           d.logger,
		MessageBusClient: mqttClient,
		BleController:    bleController,
	}
	agentService := &AgentService{
		Logger:           d.logger,
		MessageBusClient: mqttClient,
	}

	// 注入依赖到Driver
	d.BleController = bleController
	d.MessageBusClient = mqttClient
	d.CommandService = commandService
	d.AgentService = agentService

	if d.BleController == nil || d.MessageBusClient == nil {
		return fmt.Errorf("依赖未注入")
	}
	if err := d.MessageBusClient.Subscribe(TopicBLEDown, d.AgentDown); err != nil { // 转发下行数据
		d.logger.Errorf("【透明代理（↓）】 订阅下行总线失败 err: %v", err)
	}
	// 目前还未订阅过响应主题
	d.CommandService.IsSubscribeResponse = false
	return nil
}

// Discover 触发协议特定的设备发现。
func (s *Driver) Discover() error {
	return fmt.Errorf("Discover function is yet to be implemented!")

}

// ValidateDevice 校验设备协议属性。
func (s *Driver) ValidateDevice(device models.Device) error {

	protocol, ok := device.Protocols["UART"]
	if !ok {
		return errorDefault.New("Missing 'UART' protocols")
	}

	deviceLocation, ok := protocol["deviceLocation"]
	if !ok {
		return errorDefault.New("Missing 'deviceLocation' information")
	} else if deviceLocation == "" {
		return errorDefault.New("deviceLocation must not empty")
	}

	baudRate, ok := protocol["baudRate"]
	if !ok {
		return errorDefault.New("Missing 'baudRate' information")
	} else if baudRate == "" {
		return errorDefault.New("baudRate must not empty")
	}

	readTimeout, ok := protocol["readTimeout"]
	if !ok {
		return errorDefault.New("Missing 'readTimeout' information")
	} else if readTimeout == "" {
		return errorDefault.New("readTimeout must not empty")
	}
	return nil
}

// Stop 停止设备服务，释放资源。
func (d *Driver) Stop(force bool) error {
	if d.logger != nil {
		d.logger.Infof("正在停止BLE代理服务 (force=%v)", force)
	}

	// 关闭MessageBus客户端
	if d.MessageBusClient != nil {
		if closer, ok := d.MessageBusClient.(interface{ Disconnect() error }); ok {
			err := closer.Disconnect()
			if err != nil {
				d.logger.Errorf("MessageBus客户端关闭失败: %v", err)
			} else {
				d.logger.Debug("MessageBus客户端已断开连接")
			}
		} else {
			d.logger.Debug("MessageBus客户端不支持关闭操作")
		}
	}

	// 关闭BLE控制器和串口
	if d.BleController != nil {
		if closer, ok := d.BleController.(interface{ Close() error }); ok {
			err := closer.Close()
			if err != nil {
				d.logger.Errorf("BLE控制器关闭失败: %v", err)
			} else {
				d.logger.Debug("BLE控制器已关闭")
			}
		} else {
			d.logger.Debug("BLE控制器不支持关闭操作")
		}
	}

	if d.logger != nil {
		d.logger.Info("BLE代理服务已停止")
	}

	return nil
}

// AddDevice 添加设备回调函数。
func (d *Driver) AddDevice(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	d.logger.Debugf("新设备已添加: %s", deviceName)
	return nil
}

// UpdateDevice 更新设备回调函数。
func (d *Driver) UpdateDevice(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	d.logger.Debugf("设备 %s 已更新", deviceName)

	return nil
}

// RemoveDevice 移除设备回调函数。
func (d *Driver) RemoveDevice(deviceName string, protocols map[string]models.ProtocolProperties) error {
	d.logger.Debugf("设备 %s 已移除", deviceName)

	return nil
}
