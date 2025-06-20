package ble

// BLECommand 表示BLE AT命令
type BLECommand string

// BLE AT命令常量定义
// 命名规则：Command + 功能描述
const (
	// 基础控制命令
	CommandReset   BLECommand = "AT+QRST\r\n"
	CommandVersion BLECommand = "AT+QVERSION\r\n"
	CommandGetAddr BLECommand = "AT+QBLEADDR?\r\n"

	// 设备初始化命令
	CommandInitPeripheral BLECommand = "AT+QBLEINIT=2\r\n"
	CommandSetDeviceName  BLECommand = "AT+QBLENAME=QuecHCM111Z\r\n"

	// 广播相关命令
	CommandSetAdvertisingParams BLECommand = "AT+QBLEADVPARAM=150,150\r\n"
	CommandStartAdvertising     BLECommand = "AT+QBLEADVSTART\r\n"

	// GATT服务相关命令
	CommandCreateGATTService        BLECommand = "AT+QBLEGATTSSRV=fff1\r\n"
	CommandCreateGATTCharacteristic BLECommand = "AT+QBLEGATTSCHAR=fff2\r\n"
	CommandCompleteGATTService      BLECommand = "AT+QBLEGATTSSRVDONE\r\n"
)

// String 返回命令的字符串表示（用于日志和调试）
func (cmd BLECommand) String() string {
	return string(cmd)
}
