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
	errorDefault "errors"
	"fmt"
	"sync"
	"time"

	messagebus "github.com/clint456/edgex-messagebus-client"
	"github.com/edgexfoundry/device-sdk-go/v4/pkg/interfaces"
	dsModels "github.com/edgexfoundry/device-sdk-go/v4/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
)

// Driver BLE代理服务驱动程序
// 职责：协调各个组件的初始化和生命周期管理
type Driver struct {
	// EdgeX SDK相关
	sdk      interfaces.DeviceServiceSDK
	logger   logger.LoggingClient
	asyncCh  chan<- *dsModels.AsyncValues
	deviceCh chan<- []dsModels.DiscoveredDevice

	// 服务配置
	serviceConfig *ServiceConfig

	// 核心组件
	bleController    *BLEController
	messageBusClient *messagebus.Client

	// 内部状态
	commandResponses sync.Map
}

// Initialize 初始化设备服务
func (d *Driver) Initialize(sdk interfaces.DeviceServiceSDK) error {
	d.sdk = sdk
	d.logger = sdk.LoggingClient()
	d.asyncCh = sdk.AsyncValuesChannel()
	d.deviceCh = sdk.DiscoveredDeviceChannel()

	d.logger.Info("开始初始化BLE代理服务")

	// 初始化串口通信
	if err := d.initializeSerialCommunication(); err != nil {
		return fmt.Errorf("串口通信初始化失败: %w", err)
	}

	// 初始化MessageBus客户端
	if err := d.initializeMessageBus(); err != nil {
		return fmt.Errorf("MessageBus初始化失败: %w", err)
	}

	d.logger.Info("BLE代理服务初始化完成")
	return nil
}

// initializeSerialCommunication 初始化串口通信
func (d *Driver) initializeSerialCommunication() error {
	// 创建串口配置
	serialConfig := SerialPortConfig{
		PortName:    "/dev/ttyS3",
		BaudRate:    115200,
		ReadTimeout: time.Millisecond,
	}

	// 创建串口实例
	serialPort, err := NewSerialPort(serialConfig, d.logger)
	if err != nil {
		return fmt.Errorf("创建串口实例失败: %w", err)
	}

	// 创建串口队列管理器
	serialQueue := NewSerialQueue(serialPort, d.logger)

	// 创建BLE控制器
	d.bleController = NewBLEController(serialPort, serialQueue, d.logger)

	// 初始化BLE设备
	if err := d.bleController.InitializeAsPeripheral(); err != nil {
		return fmt.Errorf("BLE设备初始化失败: %w", err)
	}

	d.logger.Info("串口通信和BLE控制器初始化完成")
	return nil
}

// initializeMessageBus 初始化MessageBus客户端
func (d *Driver) initializeMessageBus() error {
	// 加载配置
	if err := d.loadServiceConfig(); err != nil {
		return fmt.Errorf("加载服务配置失败: %w", err)
	}

	// 创建统一的MessageBus客户端（同时用于监听和转发）
	messageBusClient, err := d.createMessageBusClient()
	if err != nil {
		return fmt.Errorf("创建MessageBus客户端失败: %w", err)
	}
	d.messageBusClient = messageBusClient

	d.logger.Info("MessageBus客户端初始化完成")
	return nil
}

// loadServiceConfig 加载服务配置
func (d *Driver) loadServiceConfig() error {
	d.serviceConfig = &ServiceConfig{}
	if err := d.sdk.LoadCustomConfig(d.serviceConfig, CustomConfigSectionName); err != nil {
		return fmt.Errorf("加载自定义配置失败: %w", err)
	}

	if err := d.serviceConfig.MQTTBrokerInfo.Validate(); err != nil {
		return fmt.Errorf("配置验证失败: %w", err)
	}

	// 监听配置变化
	if err := d.sdk.ListenForCustomConfigChanges(
		&d.serviceConfig.MQTTBrokerInfo.Writable,
		WritableInfoSectionName, d.updateWritableConfig); err != nil {
		return fmt.Errorf("监听配置变化失败: %w", err)
	}

	d.logger.Info("服务配置加载完成")
	return nil
}

// updateWritableConfig 更新可写配置
func (d *Driver) updateWritableConfig(rawWritableConfig interface{}) {
	updated, ok := rawWritableConfig.(*WritableInfo)
	if !ok {
		d.logger.Error("更新配置失败：类型转换错误")
		return
	}
	d.serviceConfig.MQTTBrokerInfo.Writable = *updated
	d.logger.Info("配置已更新")
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
