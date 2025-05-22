package bledriver

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"

	"github.com/edgexfoundry/device-sdk-go/v3/pkg/interfaces"
	dsModels "github.com/edgexfoundry/device-sdk-go/v3/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"
	"github.com/spf13/cast"
)

type BleDriver struct {
	sdk        interfaces.DeviceServiceSDK
	lc         logger.LoggingClient
	asyncCh    chan<- *dsModels.AsyncValues
	deviceCh   chan<- []dsModels.DiscoveredDevice
	uart       map[string]*Uart
	initSwitch bool
	sendMesg   string
}

func (s *BleDriver) Initialize(sdk interfaces.DeviceServiceSDK) error {
	s.sdk = sdk
	s.lc = sdk.LoggingClient()
	s.asyncCh = sdk.AsyncValuesChannel()
	s.deviceCh = sdk.DiscoveredDeviceChannel()

	s.uart = make(map[string]*Uart)

	return nil
}

// 在 SDK 完成初始化后，启动运行设备服务启动任务初始化。
// 这允许设备服务安全地使用 DeviceServiceSDK接口
func (s *BleDriver) Start() error {
	return nil
}

// HandleReadCommands 被指定设备的协议读取操作触发。
// HandleReadCommands triggers a protocol Read operation for the specified device.
func (s *BleDriver) HandleReadCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []dsModels.CommandRequest) (res []*dsModels.CommandValue, err error) {

	const castError = "Failed to parse %s reading: %v"
	const createCommandValueError = "Failed to create %s reading: %v"

	res = make([]*dsModels.CommandValue, len(reqs))

	var deviceLocation string
	var baudRate int

	for i, protocol := range protocols {
		deviceLocation = fmt.Sprintf("%v", protocol["deviceLocation"])
		baudRate, _ = cast.ToIntE(protocol["baudRate"])

		s.lc.Debugf("Driver.HandleReadCommands(): protocol = %v, device location = %v, baud rate = %v", i, deviceLocation, baudRate)
	}

	for i, req := range reqs {
		s.lc.Debugf("Driver.HandleReadCommands(): protocols: %v resource: %v attributes: %v", protocols, req.DeviceResourceName, req.Attributes)

		// Get the value type from device profile
		valueType := req.Type
		s.lc.Debugf("Driver.HandleReadCommands(): value type = %v", valueType)

		key_type_value := fmt.Sprintf("%v", req.Attributes["type"])

		if key_type_value == "ble" {
			key_maxbytes_value, _ := cast.ToIntE(req.Attributes["maxbytes"])
			key_timeout_value, _ := cast.ToIntE(req.Attributes["timeout"])

			// check device is already initialized
			if _, ok := s.uart[deviceLocation]; ok {
				s.lc.Debugf("Driver.HandleReadCommands(): Device %v is already initialized with baud - %v, maxbytes - %v, timeout - %v", s.uart[deviceLocation], baudRate, key_maxbytes_value, key_timeout_value)
			} else {
				// initialize device for the first time
				s.uart[deviceLocation], _ = NewUart(deviceLocation, baudRate, key_timeout_value, s.lc)
				s.uart[deviceLocation].rxbuf = nil

				s.lc.Debugf("Driver.HandleReadCommands(): Device %v initialized for the first time with baud - %v, maxbytes - %v, timeout - %v", s.uart[deviceLocation], baudRate, key_maxbytes_value, key_timeout_value)
			}

			if err := s.uart[deviceLocation].UartRead(key_maxbytes_value, s.lc); err != nil {
				return nil, fmt.Errorf("Driver.HandleReadCommands(): Reading UART failed: %v", err)
			}

			rxbuf := hex.EncodeToString(s.uart[deviceLocation].rxbuf)
			s.lc.Debugf("Driver.HandleReadCommands(): Received Data = %s", rxbuf)

			// Pass the received values to higher layers
			// Handle data based on the value type mentioned in device profile
			var cv *dsModels.CommandValue
			switch valueType {
			case common.ValueTypeInt8:
				value, err := strconv.ParseInt(rxbuf, 16, 8)
				if err != nil {
					return nil, fmt.Errorf(castError, req.DeviceResourceName, err)
				}
				cv, err = dsModels.NewCommandValue(req.DeviceResourceName, valueType, int8(value))
				if err != nil {
					return nil, fmt.Errorf(createCommandValueError, req.DeviceResourceName, err)
				}
			case common.ValueTypeInt16:
				value, err := strconv.ParseInt(rxbuf, 16, 16)
				if err != nil {
					return nil, fmt.Errorf(castError, req.DeviceResourceName, err)
				}
				cv, err = dsModels.NewCommandValue(req.DeviceResourceName, valueType, int16(value))
				if err != nil {
					return nil, fmt.Errorf(createCommandValueError, req.DeviceResourceName, err)
				}
			case common.ValueTypeString:
				cv, err = dsModels.NewCommandValue(req.DeviceResourceName, valueType, rxbuf)
				if err != nil {
					return nil, fmt.Errorf(createCommandValueError, req.DeviceResourceName, err)
				}
			default:
				return nil, fmt.Errorf("Driver.HandleReadCommands(): Unsupported value type: %v", valueType)
			}

			s.uart[deviceLocation].rxbuf = nil
			res[i] = cv
			s.lc.Debugf("Driver.HandleReadCommands(): Response = %v", res[i])
		}
	}

	return res, nil
}

// HandleWriteCommands 传递一个 CommandRequest 结构片段，每个片段代表特定设备资源的资源操作。
// 由于这些命令都是执行命令，因此 params 为每个命令提供命令参数。
func (s *BleDriver) HandleWriteCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []dsModels.CommandRequest,
	params []*dsModels.CommandValue) error {

	var deviceLocation string
	var baudRate int

	for i, protocol := range protocols {
		deviceLocation = fmt.Sprintf("%v", protocol["deviceLocation"])
		baudRate, _ = cast.ToIntE(protocol["baudRate"])

		s.lc.Debugf("BleBleDriver.HandleWriteCommands(): protocol = %v, device location = %v, baud rate = %v", i, deviceLocation, baudRate)
	}

	for i, req := range reqs {
		s.lc.Debugf("BleBleDriver.HandleWriteCommands(): deviceResourceName = %v", req.DeviceResourceName)
		s.lc.Debugf("BleBleDriver.HandleWriteCommands(): protocols: %v, resource: %v, attribute: %v, parameters: %v", protocols, req.DeviceResourceName, req.Attributes, params)

		// Get the value type from device profile
		valueType := req.Type
		s.lc.Debugf("BleBleDriver.HandleWriteCommands(): value type = %v", valueType)

		// TODO：利用工厂函数将其封装起来
		key_timeout_value, err := cast.ToIntE(req.Attributes["timeout"])
		if err != nil {
			return err
		}
		// 创建串口对象
		if _, ok := s.uart[deviceLocation]; !ok {
			s.uart[deviceLocation], err = NewUart(deviceLocation, baudRate, key_timeout_value)
			if err != nil {
				s.lc.Errorf("BleDriver.HandleWriteCommands(): 串口设备对象 创建失败：%v", err)
				return err
			}
		}
		at := NewAtCommand(s.uart[deviceLocation], s.lc) // 创建AT指令控制对象

		key_type_value := fmt.Sprintf("%v", req.Attributes["type"])
		if key_type_value == "ble" {
			switch req.DeviceResourceName {
			case "ble_init":
				if s.initSwitch, err = params[i].BoolValue(); err != nil || s.initSwitch {
					if nil != at.BleInit_2() { //初始化模式2，开启广播
						s.lc.Errorf("BleDriver.HandleWriteCommands(): BLE初始化模式2 失败：%v", err)
						return err
					}
					s.lc.Debugf("BleDriver.HandleWriteCommands(): BLE初始化模块2 成功")
				}
			case "ble_send":
				if s.sendMesg, err = params[i].StringValue(); err != nil {
					if nil != at.BleSend(s.sendMesg) { //控制BLE发出数据
						s.lc.Errorf("BleDriver.HandleWriteCommands(): BLE发出数据 失败：%v", err)
						return err
					}
					s.lc.Debugf("BleDriver.HandleWriteCommands(): BLE发出数据 成功")
				}
			}
		}

	}
	return nil
}

// Discover triggers protocol specific device discovery, asynchronously writes
// the results to the channel which is passed to the implementation via
// ProtocolDriver.Initialize()
func (s *BleDriver) Discover() error {
	return fmt.Errorf("Discover function is yet to be implemented!")
}

// ValidateDevice triggers device's protocol properties validation, returns error
// if validation failed and the incoming device will not be added into EdgeX
func (s *BleDriver) ValidateDevice(device models.Device) error {
	protocol, ok := device.Protocols["UART"]
	if !ok {
		return errors.New("Missing 'UART' protocols")
	}

	deviceLocation, ok := protocol["deviceLocation"]
	if !ok {
		return errors.New("Missing 'deviceLocation' information")
	} else if deviceLocation == "" {
		return errors.New("deviceLocation must not empty")
	}

	baudRate, ok := protocol["baudRate"]
	if !ok {
		return errors.New("Missing 'baudRate' information")
	} else if baudRate == "" {
		return errors.New("baudRate must not empty")
	}

	return nil
}

// Stop the protocol-specific DS code to shutdown gracefully, or
// if the force parameter is 'true', immediately. The BleBleDriver is responsible
// for closing any in-use channels, including the channel used to send async
// readings (if supported).
func (s *BleDriver) Stop(force bool) error {
	// Then Logging Client might not be initialized
	if s.lc != nil {
		s.lc.Debugf(fmt.Sprintf("BleBleDriver.Stop called: force=%v", force))
	}
	return nil
}

// AddDevice is a callback function that is invoked
// when a new Device associated with this Device Service is added
func (s *BleDriver) AddDevice(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	s.lc.Debugf(fmt.Sprintf("a new Device is added: %s", deviceName))
	return nil
}

// UpdateDevice is a callback function that is invoked
// when a Device associated with this Device Service is updated
func (s *BleDriver) UpdateDevice(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	s.lc.Debugf(fmt.Sprintf("Device %s is updated", deviceName))
	return nil
}

// RemoveDevice is a callback function that is invoked
// when a Device associated with this Device Service is removed
func (s *BleDriver) RemoveDevice(deviceName string, protocols map[string]models.ProtocolProperties) error {
	s.lc.Debugf(fmt.Sprintf("Device %s is removed", deviceName))
	return nil
}
