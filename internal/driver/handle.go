package driver

import (
	"fmt"

	dsModels "github.com/edgexfoundry/device-sdk-go/v4/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
)

/* Web UI 需求：
（1）支持设置蓝牙模块配置 😒
（2）支持配置【蓝牙设备】连接、断开 😒
（3）支持数据传输测试 😒
*/
// HandleReadCommands 处理读取命令。
func (d *Driver) HandleReadCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []dsModels.CommandRequest) (res []*dsModels.CommandValue, err error) {
	d.logger.Debugf("处理设备 %s 的读取命令", deviceName)
	// TODO: 实现UI具体的读取逻辑

	return nil, fmt.Errorf("读取命令暂未实现")
}

// HandleWriteCommands 处理写入命令。
func (d *Driver) HandleWriteCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []dsModels.CommandRequest, params []*dsModels.CommandValue) error {
	d.logger.Debugf("处理设备 %s 的写入命令", deviceName)
	// TODO: 实现UI具体的写入逻辑
	return nil
}
