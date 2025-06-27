package driver

import (
	"device-ble/internal/interfaces"
	"fmt"
	"log"

	blecommand "device-ble/pkg/ble"

	dsModels "github.com/edgexfoundry/device-sdk-go/v4/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
)

/* Web UI 需求：
（1）支持设置蓝牙模块配置
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
	for _, param := range params {
		if err := d.handleWrite(param, &d.BleController); err != nil {
			return err
		}
	}
	return nil
}

// handleWrite 对命令进行分类处理
func (d *Driver) handleWrite(param *dsModels.CommandValue, ble *interfaces.BLEController) error {
	switch param.DeviceResourceName {
	case "Setting&&PeripheralInit":
		objValue, err := param.ObjectValue()
		if err != nil {
			return fmt.Errorf("resourceObjectArray.write: failed to get object array value: %v", err)
		}
		cmd := objValue.(map[string]interface{})
		fmt.Printf("Setting&&PeripheralInit: %v\n", cmd)

		// 数据合法性检查
		cmdName, ok := cmd["BleName"].(string)
		if !ok {
			return fmt.Errorf("输入的BleName不是String类型")
		}
		cmdTxPower, ok := cmd["TxPower"].(int8)
		if !ok {
			return fmt.Errorf("输入的TxPower不是int8类型")
		}

		fmt.Printf("Setting&&PeripheralInit  BleName: %v, TxPower: %v\n", cmdName, cmdTxPower)

		return d.handleSettingPeripheralInit(cmdName, cmdTxPower, ble)

	case "SendString":
		{
			fmt.Printf("SendString: %v\n", param.Value)
		}
	}

	return nil
}

// 自定义初始化蓝牙模块
func (d *Driver) handleSettingPeripheralInit(BleName string, TxPower int8, ble *interfaces.BLEController) error {
	var cmds []string
	// 1. 添加通用模块控制命令
	cmds = append(cmds, blecommand.Restart()) // AT+QRST\r\n

	// 2. 添加 BLE 初始化与配置命令
	if cmd, err := blecommand.Init(2); err == nil { // Peripheral 角色
		cmds = append(cmds, cmd) // AT+QBLEINIT=1\r\n
	} else {
		log.Printf("Error generating Init: %v", err)
	}

	if cmd, err := blecommand.SetDeviceName(BleName); err == nil {
		cmds = append(cmds, cmd) // AT+QBLENAME=""\r\n
	} else {
		log.Printf("Error generating SetDeviceName: %v", err)
	}
	// 3. 添加广播控制命令
	cmds = append(cmds, blecommand.StartAdvertising()) // AT+QBLEADVSTART\r\n

	// 4. 添加 GATT 服务端命令
	if cmd, err := blecommand.AddService("180F"); err == nil { // 示例 UUID
		cmds = append(cmds, cmd) // AT+QBLEGATTSSRV="180F"\r\n
	} else {
		log.Printf("Error generating AddService: %v", err)
	}

	if cmd, err := blecommand.AddCharacteristic("2A19", 0x02|0x10); err == nil { // Read + Notify
		cmds = append(cmds, cmd) // AT+QBLEGATTSCHAR="2A19",18\r\n
	} else {
		log.Printf("Error generating AddCharacteristic: %v", err)
	}

	cmds = append(cmds, blecommand.FinishGATTServer()) // AT+QBLEGATTSSRVDONE\r\n

	if cmd, err := blecommand.SendNotify(0, 1, "1234"); err == nil { // 示例 Notify
		cmds = append(cmds, cmd) // AT+QBLEGATTSNTFY=0,1,"1234"\r\n
	} else {
		log.Printf("Error generating SendNotify: %v", err)
	}

	// 打印 cmds 切片内容
	fmt.Println("Generated AT Commands:")
	for i, cmd := range cmds {
		fmt.Printf("%d: %s", i, cmd)
	}

	return nil
}
