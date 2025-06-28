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
// 获取蓝牙的固件版本号 AT+QVERSION
// 本机 BLE MAC 地址 AT+QBLEADDR?
// 蓝牙的广播参数   AT+QBLEADVPARAM?
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
		if err := d.handleWrite(param, d.BleController); err != nil {
			return err
		}
	}
	return nil
}

// handleWrite 对命令进行分类处理
func (d *Driver) handleWrite(param *dsModels.CommandValue, ble interfaces.BLEController) error {
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
		return d.handleSetPeripheralInit(cmdName, ble)

	case "SetTxPower":
		{
			int8Value, err := param.Int8Value()
			if err != nil {
				return fmt.Errorf("resourceObjectArray.write: failed to get int8 value: %v", err)
			}
			fmt.Printf("SetTxPower: %v\n", param.Value)
			return d.handleSetTxPower(int8Value, d.BleController)
		}
	case "SetBaud":
		{
			int64Value, err := param.Int64Value()
			if err != nil {
				return fmt.Errorf("resourceObjectArray.write: failed to get int64 value: %v", err)
			}
			fmt.Printf("SetBaud: %v\n", param.Value)
			return d.handleSetBaud(int64Value, d.BleController)
		}

	case "SendString":
		{
			stringValue, err := param.StringValue()
			if err != nil {
				return fmt.Errorf("resourceObjectArray.write: failed to get int64 value: %v", err)
			}
			fmt.Printf("SendString: %v\n", param.Value)
			return d.handleSendString(stringValue, d.BleController)
		}
	}

	return nil
}

// 自定义初始化蓝牙模块
// TODO：还需要加入自定义特征值，并同步到jsonSender当中
func (d *Driver) handleSetPeripheralInit(BleName string, ble interfaces.BLEController) error {
	var cmds []string
	// 1. 添加通用模块控制命令
	cmds = append(cmds, blecommand.Restart()) // AT+QRST\r\n
	// 2. 添加 BLE 初始化与配置命令
	if cmd, err := blecommand.Init(2); err == nil { // Peripheral 角色
		cmds = append(cmds, cmd) // AT+QBLEINIT=2\r\n
	} else {
		log.Printf("Error generating Init: %v", err)
	}
	// 3. 设置BLE名称
	if cmd, err := blecommand.SetDeviceName(BleName); err == nil {
		cmds = append(cmds, cmd) // AT+QBLENAME="<BleName>"\r\n
	} else {
		log.Printf("Error generating SetDeviceName: %v", err)
	}
	// 4. 添加 GATT 服务端命令
	if cmd, err := blecommand.AddService("fff1"); err == nil { // 示例 UUID
		cmds = append(cmds, cmd) // AT+QBLEGATTSSRV=180F\r\n
	} else {
		log.Printf("Error generating AddService: %v", err)
	}
	// 5. 添加 GATT 服务端特征值
	if cmd, err := blecommand.AddCharacteristic("fff2"); err == nil { // 默认Read + Notify
		cmds = append(cmds, cmd) // AT+QBLEGATTSCHAR=2A19\r\n
	} else {
		log.Printf("Error generating AddCh aracteristic: %v", err)
	}
	// 6. 完成 GATT 服务配置
	cmds = append(cmds, blecommand.FinishGATTServer()) // AT+QBLEGATTSSRVDONE\r\n
	// 7. 开始广播
	cmds = append(cmds, blecommand.StartAdvertising()) // AT+QBLEADVSTART\r\n
	// 打印 cmds 切片内容
	// fmt.Println("Generated AT Commands:")
	for i, cmd := range cmds {
		fmt.Printf("%d: %s", i, cmd)
	}

	return ble.CustomInitializeBle(cmds)
}

func (d *Driver) handleSetTxPower(TxPower int8, ble interfaces.BLEController) error {
	if cmd, err := blecommand.SetTxPower(TxPower); err != nil {
		return fmt.Errorf("Error generating SetBaud: %v", err)
	} else {
		return ble.SendSingle(cmd)
	}

}

func (d *Driver) handleSetBaud(Baud int64, ble interfaces.BLEController) error {
	if cmd, err := blecommand.SetBaud(Baud); err != nil {
		return fmt.Errorf("Error generating SetBaud: %v", err)

	} else {
		return ble.SendSingle(cmd)
	}

}

func (d *Driver) handleSendString(Str string, ble interfaces.BLEController) error {
	if len(Str) > 223 {
		return fmt.Errorf("Error, Becuase sending str out of range (240bit)")
	}
	if cmd, err := blecommand.SendNotify("fff2", Str); err != nil {
		return fmt.Errorf("Error generating SendString")
	} else {
		return ble.SendSingle(cmd)
	}

}
