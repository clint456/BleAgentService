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

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/edgexfoundry/device-sdk-go/v4/pkg/interfaces"
	dsModels "github.com/edgexfoundry/device-sdk-go/v4/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
	"github.com/edgexfoundry/go-mod-messaging/v4/messaging"
)

type Driver struct {
	sdk              interfaces.DeviceServiceSDK
	lc               logger.LoggingClient
	asyncCh          chan<- *dsModels.AsyncValues
	deviceCh         chan<- []dsModels.DiscoveredDevice
	AsyncCh          chan<- *dsModels.AsyncValues
	serviceConfig    *ServiceConfig
	mqttClient       mqtt.Client             // 监听客户端
	transmitClient   messaging.MessageClient //转发客户端
	CommandResponses sync.Map
	ble              *BleController
}

// Initialize performs protocol-specific initialization for the device
// service.
func (s *Driver) Initialize(sdk interfaces.DeviceServiceSDK) error {

	s.sdk = sdk
	s.lc = sdk.LoggingClient()
	s.asyncCh = sdk.AsyncValuesChannel()
	s.deviceCh = sdk.DiscoveredDeviceChannel()

	serial, err := NewSerialPort("/dev/ttyS3", 115200)
	if err != nil {
		s.lc.Errorf("❌ 串口初始化失败: %v", err)
		return err
	}

	s.ble = NewBleController(serial, true)
	if err := s.ble.InitAsPeripheral(); err != nil {
		s.lc.Errorf("❌ BLE 初始化失败: %v", err)
		return err
	}
	s.lc.Debugf("✅ BLE 初始化为外围设备完成！")

	if err := s.initialMqttClient(); err != nil {
		s.lc.Errorf("❌ MQTT 初始化失败: %v", err)
		return err
	}

	return nil
}

// Start runs device service startup tasks after the SDK has been completely
// initialized. This allows device service to safely use DeviceServiceSDK
// interface features in this function call
func (s *Driver) Start() error {
	return nil
}

// HandleReadCommands triggers a protocol Read operation for the specified device.
func (s *Driver) HandleReadCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []dsModels.CommandRequest) (res []*dsModels.CommandValue, err error) {
	s.lc.Debug("📖 Read 操作被触发")
	return nil, fmt.Errorf("👋 hello")
}

// HandleWriteCommands passes a slice of CommandRequest struct each representing
// a ResourceOperation for a specific device resource.
// Since the commands are actuation commands, params provide parameters for the individual
// command.
func (s *Driver) HandleWriteCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []dsModels.CommandRequest,
	params []*dsModels.CommandValue) error {
	s.lc.Debug("✍️ 写操作被触发")
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

// Stop the protocol-specific DS code to shutdown gracefully, or
// if the force parameter is 'true', immediately. The driver is responsible
// for closing any in-use channels, including the channel used to send async
// readings (if supported).
func (s *Driver) Stop(force bool) error {
	// Then Logging Client might not be initialized
	if s.lc != nil {
		s.lc.Debugf(fmt.Sprintf("Driver.Stop called: force=%v", force))
	}
	s.ble.serial.Close()
	return nil
}

// AddDevice is a callback function that is invoked
// when a new Device associated with this Device Service is added
func (s *Driver) AddDevice(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	s.lc.Debugf(fmt.Sprintf("a new Device is added: %s", deviceName))
	return nil
}

// UpdateDevice is a callback function that is invoked
// when a Device associated with this Device Service is updated
func (s *Driver) UpdateDevice(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	s.lc.Debugf(fmt.Sprintf("Device %s is updated", deviceName))
	return nil
}

// RemoveDevice is a callback function that is invoked
// when a Device associated with this Device Service is removed
func (s *Driver) RemoveDevice(deviceName string, protocols map[string]models.ProtocolProperties) error {
	s.lc.Debugf(fmt.Sprintf("Device %s is removed", deviceName))
	return nil
}
