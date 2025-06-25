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
	internalif "device-ble/internal/interfaces"
	"log"
	"time"

	"device-ble/pkg/ble"
	"device-ble/pkg/dataparse"
	"device-ble/pkg/mqttbus"
	"device-ble/pkg/uart"
	errorDefault "errors"
	"fmt"
	"sync"

	edgexif "github.com/edgexfoundry/device-sdk-go/v4/pkg/interfaces"
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
	sdk      edgexif.DeviceServiceSDK
	logger   logger.LoggingClient
	asyncCh  chan<- *dsModels.AsyncValues
	deviceCh chan<- []dsModels.DiscoveredDevice
	// 服务配置
	Config internalif.ConfigProvider

	// 核心组件
	BleController    internalif.BLEController
	MessageBusClient internalif.MessageBusClient

	// 内部状态
	commandResponses sync.Map
}

// Initialize 初始化设备服务
func (d *Driver) Initialize(sdk edgexif.DeviceServiceSDK) error {
	d.sdk = sdk
	d.logger = sdk.LoggingClient()
	d.asyncCh = sdk.AsyncValuesChannel()
	d.deviceCh = sdk.DiscoveredDeviceChannel()

	// 1. 读取配置（此处假设已通过 d.config 注入或可直接 new）
	if d.Config == nil {
		return fmt.Errorf("ConfigProvider 未注入")
	}

	// 2. 初始化串口、BLE 控制器
	serialCfg := d.Config.GetSerialConfig()
	serialPort, err := uart.NewSerialPort(serialCfg, d.logger)
	if err != nil {
		return fmt.Errorf("创建串口实例失败: %w", err)
	}

	serialQueue := uart.NewSerialQueue(serialPort, d.logger,
		// 闭包注入回调方法，安全共享driver对象资源
		d.HandleUpCommandCallback,
		d.HandleUpAgentCallback,
	)
	d.BleController = ble.NewBLEController(serialPort, serialQueue, d.logger)
	if err := d.BleController.InitializeAsPeripheral(); err != nil {
		return fmt.Errorf("BLE设备初始化失败: %w", err)
	}

	// 3. 初始化 messageBusClient
	mqttClientConfig := d.Config.GetMQTTConfig()
	d.MessageBusClient, err = mqttbus.NewEdgexMessageBusClient(mqttClientConfig, d.logger)
	if err != nil {
		return fmt.Errorf("MessageBusClient 创建失败")
	}

	d.logger.Info("BLE代理服务初始化完成")
	// 监听消息总线是否有透明代理下发命令
	// TODO 透明代理消息数据解析
	/* 下发数据格式：
	payload{
	timestamp:"",  //  Unix纳秒时间戳
	data:""   //原始数据
	}
	*/
	// 测试发送连通性
	d.MessageBusClient.Subscribe("edgex/service/data/device_ble/dwon", d.agentDown)
	return nil
}

func (d *Driver) HandleUpAgentCallback(data string) {
	if data == "" {
		return
	}
	type Payload struct {
		Timestamp int64  // Unix 纳秒时间戳
		Data      string // 原始数据（例如串口内容）
	}
	fmt.Printf("透明代理：收到上行数据: %s\n", data)
	// 解析上报数据
	p := Payload{
		Timestamp: time.Now().UnixNano(), // 当前时间戳（纳秒）
		Data:      data,
	}
	// 转发至MessageBus
	if d.MessageBusClient != nil {
		err := d.MessageBusClient.Publish("edgex/service/data/device_ble/up", p)
		if err != nil {
			fmt.Printf("【透明代理上行】转发至消息总线失败 ❌: %v \n", err) // 记录错误日志
		} else {
			fmt.Printf("【透明代理上行】转发至消息总线成功 ✔ \n") // 记录错误日志
		}
	}
	return
}

func (d *Driver) HandleUpCommandCallback(cmd string) {
	fmt.Printf("收到控制上报命令: %s\n", cmd)
	// TODO: 解析运维系统命令并响应
	/* 上报命令格式
	Payload{
		statusCode:"0",     // 状态码，默认为0，异常为非0
		timestamp:"",           //  Unix纳秒时间戳
		commandType:"allstatus",   // 命令类型：allstatus/monitor
		commandParam:" ",  // 留空
	}
	*/

	/* 解析上报命令
	1. 检查statusCode字段
	2. 判断commandType字段，分发allstatus、monitor协程任务
	*/

	/* allstatus任务
	1. 订阅系统事件
	2. 获取status字段以及设备相对应字段，json格式化，交由ai处理,整理发送数据包
	3. ataparse.SendToBlE发送给设备作为响应
	*/

	/* monitor 协程任务
	1. 判断上一个监听是否结束
	2. 订阅指定设备事件
	3. 获取响应字段，格式化，整理发送数据包
	4. dataparse.SendToBlE发送给设备作为响应
	*/

	switch cmd {
	case "allstatus":
		fmt.Printf("【运维】开始查询所有设备状态")
		err := d.MessageBusClient.SubscribeResponse("edgex/response/core-command/#")
		if err != nil {
			log.Fatalf("订阅响应失败: %v", err)
		}
		// 发送请求
		payload := ""
		resp, err := d.MessageBusClient.Request("edgex/core/commandquery/request/all", payload)
		if err != nil {
			log.Fatalf("请求失败: %v", err)
		}
		fmt.Printf("请求系统响应： %v", resp)

	case "status":

	}

}

// Start 启动设备服务
func (d *Driver) Start() error {

	return nil
}

// agentDown 该回调函数被消息总线接收协程所调用
// 用于处理蓝牙透明代理下行数据
func (d *Driver) agentDown(topic string, envelope types.MessageEnvelope) error {
	dataparse.SendToBlE(d.BleController, envelope)
	return nil
}

// HandleReadCommands 处理读取命令
func (d *Driver) HandleReadCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []dsModels.CommandRequest) (res []*dsModels.CommandValue, err error) {
	d.logger.Debugf("处理设备 %s 的读取命令", deviceName)
	// TODO: 实现UI具体的读取逻辑
	return nil, fmt.Errorf("读取命令暂未实现")
}

// HandleWriteCommands 处理写入命令
func (d *Driver) HandleWriteCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []dsModels.CommandRequest, params []*dsModels.CommandValue) error {
	d.logger.Debugf("处理设备 %s 的写入命令", deviceName)
	// TODO: 实现UI具体的写入逻辑
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
	if d.MessageBusClient != nil {
		// 若接口有 Disconnect 方法则调用，否则跳过
		// d.messageBusClient.Disconnect()
		d.logger.Debug("MessageBus客户端已断开连接")
	}

	// 关闭BLE控制器和串口
	if d.BleController != nil {
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
