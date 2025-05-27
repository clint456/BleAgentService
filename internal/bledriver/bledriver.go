package bledriver

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/edgexfoundry/device-sdk-go/v3/pkg/interfaces"
	dsModels "github.com/edgexfoundry/device-sdk-go/v3/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"
	"github.com/spf13/cast"
)

type BleDriver struct {
	sdk            interfaces.DeviceServiceSDK
	lc             logger.LoggingClient
	asyncCh        chan<- *dsModels.AsyncValues
	deviceCh       chan<- []dsModels.DiscoveredDevice
	uart           map[string]*Uart
	initSwitch     bool
	sendMesg       string
	deviceLocation string
	baudRate       int
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
func (s *BleDriver) HandleReadCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []dsModels.CommandRequest) (res []*dsModels.CommandValue, err error) {
	var responses = make([]*dsModels.CommandValue, len(reqs)) //创建命令切片
	for i, protocol := range protocols {
		s.deviceLocation = fmt.Sprintf("%v", protocol["deviceLocation"])
		s.baudRate, _ = cast.ToIntE(protocol["baudRate"])
		s.lc.Debugf("Driver.HandleReadCommands(): protocol = %v, device location = %v, baud rate = %v", i, s.deviceLocation, s.baudRate)
	}

	for i, req := range reqs {
		s.lc.Debugf("Driver.HandleReadCommands(): protocols: %v resource: %v attributes: %v", protocols, req.DeviceResourceName, req.Attributes)
		resource, ok := s.sdk.DeviceResource(deviceName, req.DeviceResourceName)
		if !ok {
			return responses, fmt.Errorf("handle read commands failed: Device Resource %s not found", req.DeviceResourceName)
		}

		res, err := s.handleReadCommandRequest(req, resource)
		if err != nil {
			s.lc.Errorf("Handle read commands failed: %v", err)
			return responses, err
		}

		responses[i] = res
	}

	return responses, err

}

func (s *BleDriver) handleReadCommandRequest(req dsModels.CommandRequest, resource models.DeviceResource) (*dsModels.CommandValue, error) {
	var result *dsModels.CommandValue
	var err error
	const createCommandValueError = "Failed to create %s reading: %v"

	// Get the value type from device profile
	valueType := req.Type
	s.lc.Debugf("Driver.HandleReadCommands(): value type = %v", valueType)

	key_type_value := fmt.Sprintf("%v", req.Attributes["type"])

	if key_type_value == "ble" {
		key_maxbytes_value, _ := cast.ToIntE(req.Attributes["maxbytes"])
		key_timeout_value, _ := cast.ToIntE(req.Attributes["timeout"])

		// check device is already initialized
		if _, ok := s.uart[s.deviceLocation]; ok {
			s.lc.Debugf("Driver.HandleReadCommands(): Device %v is already initialized with baud - %v, maxbytes - %v, timeout - %v", s.uart[s.deviceLocation], s.baudRate, key_maxbytes_value, key_timeout_value)

		} else {
			// initialize device for the first time
			s.uart[s.deviceLocation], _ = NewUart(s.deviceLocation, s.baudRate, key_timeout_value)
			s.lc.Debugf("Driver.HandleReadCommands(): Device %v initialized for the first time with baud - %v, maxbytes - %v, timeout - %v", s.uart[s.deviceLocation], s.baudRate, key_maxbytes_value, key_timeout_value)
		}
		// 清空当前接收缓存区
		s.uart[s.deviceLocation].rxbuf = nil
		// 读取缓存区
		if err := s.uart[s.deviceLocation].UartRead(key_maxbytes_value); err != nil {
			return nil, fmt.Errorf("Driver.HandleReadCommands(): Reading UART failed: %v", err)
		}

		// Pass the received values to higher layers
		// Handle data based on the value type mentioned in device profile
		var cv *dsModels.CommandValue

		switch valueType {
		case common.ValueTypeString:
			// 字符串接收
			rxbuf := string(s.uart[s.deviceLocation].rxbuf)
			cv, err = dsModels.NewCommandValue(req.DeviceResourceName, valueType, rxbuf)
			if err != nil {
				return nil, fmt.Errorf(createCommandValueError, req.DeviceResourceName, err)
			}
		case common.ValueTypeBool:
			//获取当前蓝牙设备状态
			sta, _ := CheckAtState(s.uart[s.deviceLocation])
			cv, err = dsModels.NewCommandValue(req.DeviceResourceName, "String", string(sta))
			if err != nil {
				return nil, fmt.Errorf(createCommandValueError, req.DeviceResourceName, err)
			}
		case common.ValueTypeObject:
			//JSON数据解析
			var response map[string]interface{}
			err = json.Unmarshal(s.uart[s.deviceLocation].rxbuf, &response)
			if err != nil {
				s.lc.Errorf("Error unmarshaling response: %s", err)
				return nil, fmt.Errorf("Error unmarshaling response: %s", err)
			}
			cv, err = dsModels.NewCommandValue(req.DeviceResourceName, valueType, response)
			if err != nil {
				return nil, fmt.Errorf(createCommandValueError, req.DeviceResourceName, err)
			}
		default:
			return nil, fmt.Errorf("Driver.HandleReadCommands(): Unsupported value type: %v", valueType)
		}

		s.uart[s.deviceLocation].rxbuf = nil
		result = cv
		s.lc.Debugf("Driver.HandleReadCommands(): Response = %v", result)
	}
	return result, nil
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

		key_timeout_value, err := cast.ToIntE(req.Attributes["timeout"])
		if err != nil {
			return err
		}
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
				if s.initSwitch, err = params[i].BoolValue(); err != nil {
					s.lc.Errorf("BleDriver.HandleWriteCommands(): 获取开关失败%v", err)
				}
				if s.initSwitch {
					if nil != at.BleInit_2() { //初始化模式2，开启广播
						s.lc.Errorf("BleDriver.HandleWriteCommands(): BLE初始化模式2 失败：%v", err)
						return err
					}
					s.lc.Debugf("BleDriver.HandleWriteCommands(): BLE初始化模块2 成功")
				} else {
					// TODO关闭蓝牙模块
					s.lc.Debugf("BleDriver.HandleWriteCommands(): BLE关闭")
				}

			case "ble_data":
				if s.sendMesg, err = params[i].StringValue(); err != nil {
					s.lc.Errorf("BleDriver.HandleWriteCommands(): 获取发送的消息格式非法 ：%v", err)
				}
				if nil != at.BleSend(s.sendMesg) { //控制BLE发出数据
					s.lc.Errorf("BleDriver.HandleWriteCommands(): BLE发出数据 失败：%v", err)
					return err
				}
				s.lc.Debugf("BleDriver.HandleWriteCommands(): BLE发出数据 成功")

			}
		}

	}
	return nil
}

func newResult(resource models.DeviceResource, reading interface{}) (*dsModels.CommandValue, error) {
	var err error
	var result = &dsModels.CommandValue{}
	castError := "fail to parse %v reading, %v"

	valueType := resource.Properties.ValueType

	if !checkValueInRange(valueType, reading) {
		err = fmt.Errorf("parse reading fail. Reading %v is out of the value type(%v)'s range", reading, valueType)
		return result, err
	}

	var val interface{}
	switch valueType {
	case common.ValueTypeBool:
		val, err = cast.ToBoolE(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, resource.Name, err)
		}
	case common.ValueTypeString:
		val, err = cast.ToStringE(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, resource.Name, err)
		}
	case common.ValueTypeUint8:
		val, err = cast.ToUint8E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, resource.Name, err)
		}
	case common.ValueTypeUint16:
		val, err = cast.ToUint16E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, resource.Name, err)
		}
	case common.ValueTypeUint32:
		val, err = cast.ToUint32E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, resource.Name, err)
		}
	case common.ValueTypeUint64:
		val, err = cast.ToUint64E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, resource.Name, err)
		}
	case common.ValueTypeInt8:
		val, err = cast.ToInt8E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, resource.Name, err)
		}
	case common.ValueTypeInt16:
		val, err = cast.ToInt16E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, resource.Name, err)
		}
	case common.ValueTypeInt32:
		val, err = cast.ToInt32E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, resource.Name, err)
		}
	case common.ValueTypeInt64:
		val, err = cast.ToInt64E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, resource.Name, err)
		}
	case common.ValueTypeFloat32:
		val, err = cast.ToFloat32E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, resource.Name, err)
		}
	case common.ValueTypeFloat64:
		val, err = cast.ToFloat64E(reading)
		if err != nil {
			return nil, fmt.Errorf(castError, resource.Name, err)
		}
	case common.ValueTypeObject:
		val = reading
	default:
		return nil, fmt.Errorf("return result fail, none supported value type: %v", valueType)

	}

	result, err = dsModels.NewCommandValue(resource.Name, valueType, val)
	if err != nil {
		return nil, err
	}
	result.Origin = time.Now().UnixNano()

	return result, nil
}

func newCommandValue(valueType string, param *dsModels.CommandValue) (interface{}, error) {
	var commandValue interface{}
	var err error
	switch valueType {
	case common.ValueTypeBool:
		commandValue, err = param.BoolValue()
	case common.ValueTypeString:
		commandValue, err = param.StringValue()
	case common.ValueTypeUint8:
		commandValue, err = param.Uint8Value()
	case common.ValueTypeUint16:
		commandValue, err = param.Uint16Value()
	case common.ValueTypeUint32:
		commandValue, err = param.Uint32Value()
	case common.ValueTypeUint64:
		commandValue, err = param.Uint64Value()
	case common.ValueTypeInt8:
		commandValue, err = param.Int8Value()
	case common.ValueTypeInt16:
		commandValue, err = param.Int16Value()
	case common.ValueTypeInt32:
		commandValue, err = param.Int32Value()
	case common.ValueTypeInt64:
		commandValue, err = param.Int64Value()
	case common.ValueTypeFloat32:
		commandValue, err = param.Float32Value()
	case common.ValueTypeFloat64:
		commandValue, err = param.Float64Value()
	case common.ValueTypeObject:
		commandValue, err = param.ObjectValue()
	default:
		err = fmt.Errorf("fail to convert param, none supported value type: %v", valueType)
	}

	return commandValue, err
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
