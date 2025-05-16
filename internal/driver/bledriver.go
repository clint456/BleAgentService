package BleBleDriver

import (
	"fmt"

	"github.com/edgexfoundry/device-sdk-go/v3/pkg/interfaces"
	dsModels "github.com/edgexfoundry/device-sdk-go/v3/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"
	"github.com/spf13/cast"
)

type BleDriver struct {
	sdk      interfaces.DeviceServiceSDK
	lc       logger.LoggingClient
	asyncCh  chan<- *dsModels.AsyncValues
	deviceCh chan<- []dsModels.DiscoveredDevice
	uart     map[string]*Uart
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

func (s *BleDriver) start() error {
	return nil
}

// HandleReadCommands 被指定设备的协议读取操作触发。
func (s *BleDriver) HandleReadCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []dsModels.CommandRequest) (res []*dsModels.CommandValue, err error) {

	const castError = "Failed to parse %s reading: %v"
	const createCommandValueError = "Failed to create %s reading: %v"

	res = make([]*dsModels.CommandValue, len(reqs))

	var deviceLocation string
	var baudRate int

	for i, protocol := range protocols {
		deviceLocation = fmt.Sprintf("%v", protocol["deviceLocation"])
		baudRate, _ = cast.ToIntE(protocol["baudRate"])

		s.lc.Debugf("BleBleDriver.HandleReadCommands(): protocol = %v, device location = %v, baud rate = %v", i, deviceLocation, baudRate)
	}

	for i, req := range reqs {
		s.lc.Debugf("BleBleDriver.HandleReadCommands(): protocols: %v resource: %v attributes: %v", protocols, req.DeviceResourceName, req.Attributes)

		// Get the value type from device profile
		valueType := req.Type
		s.lc.Debugf("BleBleDriver.HandleReadCommands(): value type = %v", valueType)

		key_type_value := fmt.Sprintf("%v", req.Attributes["type"])

		if key_type_value == "ble" {
			//TODO

		}

	}
	return nil, nil
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

		var value interface{}
		var err error

		key_type_value := fmt.Sprintf("%v", req.Attributes["type"])
		if key_type_value == "ble" {
			// bool: 控制蓝牙开启关闭
			// string: 控制蓝牙发送/接收
			switch valueType {
			case common.ValueTypeBool:
				value, err = params[i].BoolValue()
			case common.ValueTypeString:
				value, err = params[i].StringValue()
			default:
				return fmt.Errorf("BleBleDriver.HandleWriteCommands(): Unsupported value type: %v", valueType)

			}
			if err != nil {
				return err
			}
			s.lc.Debugf("BleDriver.HandleWriteCommands(): %s= %v", valueType, value)
			//反射接口转化为具体数据类型值
			key_value, err := cast.ToStringE(req.DeviceResourceName)
			key_timeout_value, err := cast.ToIntE(req.Attributes["timeout"])
			if err != nil {
				return err
			}
			//判断蓝牙是否初始化，否则进行初始化
			if(s.uart)
			// 判断是上面设备操作类型
			switch key_value {
			case "ble_send":
				s.ble_send()
			case "ble_receive":
				s.ble_receive()
		}

	}
	return nil
}
func (s *BleDriver) ble_state() bool,error{

	return true, nil
}
func (s *BleDriver) ble_init() error {

	return nil
}

func (s *BleDriver) ble_send() error {

	return nil
}

func (s *BleDriver) ble_receive() error {

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
