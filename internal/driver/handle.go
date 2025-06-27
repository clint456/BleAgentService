package driver

import (
	"encoding/json"
	"fmt"

	dsModels "github.com/edgexfoundry/device-sdk-go/v4/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
)

/* Web UI 需求：
（1）支持设置蓝牙模块配置
（2）支持配置【蓝牙设备】连接、断开
*/
// HandleReadCommands 处理读取命令。
func (d *Driver) HandleReadCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []dsModels.CommandRequest) (res []*dsModels.CommandValue, err error) {
	d.logger.Debugf("处理设备 %s 的读取命令", deviceName)
	// TODO: 实现UI具体的读取逻辑
	fmt.Printf("deviceName: \n\t%s,\nprotocols: \n\t%v,\n reqs:\t%v,\n ", deviceName, protocols, reqs)
	res = make([]*dsModels.CommandValue, len(reqs))

	for i, req := range reqs {
		if dr, ok := d.sdk.DeviceResource(deviceName, req.DeviceResourceName); ok {
			fmt.Printf("第 %d 个 dsresource: \n\t%v\n ", i, dr)

		} else {
			return nil, fmt.Errorf("cannot find device resource %s from device %s in cache", req.DeviceResourceName, deviceName)
		}
	}
	return nil, fmt.Errorf("读取命令暂未实现")
}

// HandleWriteCommands 处理写入命令。
func (d *Driver) HandleWriteCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []dsModels.CommandRequest, params []*dsModels.CommandValue) error {
	d.logger.Debugf("处理设备 %s 的写入命令", deviceName)
	// TODO: 实现UI具体的写入逻辑
	fmt.Printf("deviceName: \n\t%s,\n protocols: \n\t%v,\n reqs:\t%v,\n params:\n\t%v\n", deviceName, protocols, reqs, params)
	for i, req := range reqs {
		if dr, ok := d.sdk.DeviceResource(deviceName, req.DeviceResourceName); ok {

			fmt.Printf("第 %d 个 dsresource: \n\t%v\n ", i, dr)
		} else {
			return fmt.Errorf("cannot find device resource %s from device %s in cache", req.DeviceResourceName, deviceName)
		}
	}
	return nil
}

// 处理初始化任务
func (d *Driver) handle(param *dsModels.CommandValue) error {
	objArrayValue, err := param.ObjectArrayValue()
	if err != nil {
		return fmt.Errorf("resourceObjectArray.write: failed to get object array value: %v", err)
	}

	jsonBytes, err := json.Marshal(objArrayValue)
	if err != nil {
		return fmt.Errorf("resourceObjectArray.write: failed to marshal object array value: %v", err)
	}
}
