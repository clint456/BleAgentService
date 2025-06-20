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
	"device-ble/driver/ble"
	"device-ble/driver/dataparse"
	"device-ble/driver/mqttbus"
	"device-ble/driver/uart"
	"encoding/json"
	errorDefault "errors"
	"fmt"
	"sync"

	"github.com/edgexfoundry/device-sdk-go/v4/pkg/interfaces"
	dsModels "github.com/edgexfoundry/device-sdk-go/v4/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
	"github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"
)

// Driver BLE代理服务驱动程序
// 职责：协调各个组件的初始化和生命周期管理
// 面向对象重构：所有依赖通过字段注入，便于测试和扩展
type Driver struct {
	// EdgeX SDK相关
	sdk      interfaces.DeviceServiceSDK
	logger   logger.LoggingClient
	asyncCh  chan<- *dsModels.AsyncValues
	deviceCh chan<- []dsModels.DiscoveredDevice

	// 服务配置
	serviceConfig *mqttbus.ServiceConfig
	config        ConfigProvider // 新增：配置接口注入

	// 核心组件
	bleController    *ble.BLEController
	messageBusClient mqttbus.MessageBusClient // 改为接口类型

	// 内部状态
	commandResponses sync.Map
}

// Initialize 初始化设备服务
func (d *Driver) Initialize(sdk interfaces.DeviceServiceSDK) error {
	d.sdk = sdk
	d.logger = sdk.LoggingClient()
	d.asyncCh = sdk.AsyncValuesChannel()
	d.deviceCh = sdk.DiscoveredDeviceChannel()

	// 1. 读取配置（此处假设已通过 d.config 注入或可直接 new）
	if d.config == nil {
		d.config = &Config{
			Serial: SerialConfig{
				PortName:    "/dev/ttyS3",
				BaudRate:    115200,
				ReadTimeout: 100,
			},
			MQTT: MQTTConfig{
				Host:     "localhost",
				Port:     1883,
				Protocol: "tcp",
				ClientID: "device-ble",
				QoS:      1,
				Username: "",
				Password: "",
			},
		}
	}

	// 2. 初始化串口、BLE 控制器
	serialCfg := d.config.GetSerialConfig()
	serialPort, err := uart.NewSerialPort(uart.SerialPortConfig{
		PortName:    serialCfg.PortName,
		BaudRate:    serialCfg.BaudRate,
		ReadTimeout: serialCfg.ReadTimeout,
	}, d.logger)
	if err != nil {
		return fmt.Errorf("创建串口实例失败: %w", err)
	}
	serialQueue := uart.NewSerialQueue(serialPort, d.logger)
	d.bleController = ble.NewBLEController(serialPort, serialQueue, d.logger)
	if err := d.bleController.InitializeAsPeripheral(); err != nil {
		return fmt.Errorf("BLE设备初始化失败: %w", err)
	}

	// 3. 组装 handler（闭包，注入 messageBusClient、bleController）
	handler := func(topic string, envelope types.MessageEnvelope) error {
		var data map[string]interface{}
		if err := json.Unmarshal(envelope.Payload.([]byte), &data); err != nil {
			d.logger.Errorf("解析消息失败: %v", err)
			return err
		}
		// 发布到 MessageBus
		if err := dataparse.PublishToMessageBus(d.messageBusClient, data, topic); err != nil {
			d.logger.Errorf("转发到MessageBus失败: %v", err)
			return err
		}
		// 发送到 BLE
		dataparse.SendToBlE(d.bleController, data)
		return nil
	}

	// 4. 初始化 messageBusClient，并将 handler 传递给 mqttbus
	mqttCfg := d.config.GetMQTTConfig()
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
	msgBus, err := mqttbus.NewEdgexMessageBusClient(cfgMap, d.logger, subscribeTopics, handler)
	if err != nil {
		return fmt.Errorf("MessageBus初始化失败: %w", err)
	}
	d.messageBusClient = msgBus // 类型一致

	d.logger.Info("BLE代理服务初始化完成")
	return nil
}

// Start 启动设备服务
func (d *Driver) Start() error {
	d.logger.Info("BLE代理服务已启动")
	return nil
}

// HandleReadCommands 处理读取命令
func (d *Driver) HandleReadCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []dsModels.CommandRequest) (res []*dsModels.CommandValue, err error) {
	d.logger.Debugf("处理设备 %s 的读取命令", deviceName)
	// TODO: 实现具体的读取逻辑
	return nil, fmt.Errorf("读取命令暂未实现")
}

// HandleWriteCommands 处理写入命令
func (d *Driver) HandleWriteCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []dsModels.CommandRequest, params []*dsModels.CommandValue) error {
	d.logger.Debugf("处理设备 %s 的写入命令", deviceName)
	// TODO: 实现具体的写入逻辑
	return nil
}

// Discover triggers protocol specific device discovery, asynchronously writes
// the results to the channel which is passed to the implementation via
// ProtocolDriver.Initialize()
func (s *Driver) Discover() error {
	return fmt.Errorf("Discover function is yet to be implemented!")

}

// ValidateDevice triggers device's protocol properties validation, returns error
// if validation failed and the incoming device will not be added into EdgeX
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

	return nil
}

// Stop 停止设备服务
func (d *Driver) Stop(force bool) error {
	if d.logger != nil {
		d.logger.Infof("正在停止BLE代理服务 (force=%v)", force)
	}

	// 关闭MessageBus客户端
	if d.messageBusClient != nil {
		d.messageBusClient.Disconnect()
		d.logger.Debug("MessageBus客户端已断开连接")
	}

	// 关闭BLE控制器和串口
	if d.bleController != nil {
		// TODO: 添加BLE控制器的关闭方法
		d.logger.Debug("BLE控制器已关闭")
	}

	if d.logger != nil {
		d.logger.Info("BLE代理服务已停止")
	}

	return nil
}

// AddDevice 添加设备回调函数
func (d *Driver) AddDevice(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	d.logger.Debugf("新设备已添加: %s", deviceName)
	return nil
}

// UpdateDevice 更新设备回调函数
func (d *Driver) UpdateDevice(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	d.logger.Debugf("设备 %s 已更新", deviceName)
	return nil
}

// RemoveDevice 移除设备回调函数
func (d *Driver) RemoveDevice(deviceName string, protocols map[string]models.ProtocolProperties) error {
	d.logger.Debugf("设备 %s 已移除", deviceName)
	return nil
}
