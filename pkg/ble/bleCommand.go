package ble

import (
	"fmt"
)

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

// --- 通用模块控制 ---

// Restart 生成模块重启的 AT 命令
func Restart() string {
	return "AT+QRST\r\n"
}

// GetVersion 生成查询固件版本的 AT 命令
func GetVersion() string {
	return "AT+QVERSION\r\n"
}

// SetBaud 生成设置串口波特率的 AT 命令
func SetBaud(baud int64) (string, error) {
	// 验证波特率有效性
	validBauds := map[int64]bool{9600: true, 19200: true, 38400: true, 57600: true, 115200: true, 230400: true, 460800: true, 921600: true}
	if !validBauds[baud] {
		return "", fmt.Errorf("invalid baud rate: %d, supported: %v", baud, validBauds)
	}
	return fmt.Sprintf("AT+QSETBAUD=%d\r\n", baud), nil
}

// SetTxPower 生成发送功率，限制值的大小
func SetTxPower(txpower int8) (string, error) {
	if txpower > 10 || txpower < -16 {
		return "", fmt.Errorf("txpower setting value out of range [-16,10]")
	}
	return fmt.Sprintf("AT+QTXPOWER=%d\r\n", txpower), nil
}

// --- BLE 初始化与配置 ---

// Init 生成初始化 BLE 栈的 AT 命令
func Init(role int) (string, error) {
	// 验证角色有效性
	validRoles := map[int]bool{1: true, 2: true, 4: true} // Peripheral=1, Central=2, Multi-role=4
	if !validRoles[role] {
		return "", fmt.Errorf("invalid BLE role: %d, supported: %v", role, validRoles)
	}
	return fmt.Sprintf("AT+QBLEINIT=%d\r\n", role), nil
}

// SetDeviceName 生成设置设备名称的 AT 命令
func SetDeviceName(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("device name cannot be empty")
	}
	return fmt.Sprintf("AT+QBLENAME=%s\r\n", name), nil
}

// QueryAddress 生成查询 BLE MAC 地址的 AT 命令
func QueryAddress() string {
	return "AT+QBLEADDR?\r\n"
}

// --- 广播控制 ---

// StartAdvertising 生成启动广播的 AT 命令
func StartAdvertising() string {
	return "AT+QBLEADVSTART\r\n"
}

// StopAdvertising 生成停止广播的 AT 命令
func StopAdvertising() string {
	return "AT+QBLEADVSTOP\r\n"
}

// --- GATT 服务端 ---

// AddService 生成添加 GATT 服务的 AT 命令
func AddService(uuid string) (string, error) {
	if uuid == "" {
		return "", fmt.Errorf("UUID cannot be empty")
	}
	return fmt.Sprintf("AT+QBLEGATTSSRV=%s\r\n", uuid), nil
}

// AddCharacteristic 生成添加特征值的 AT 命令
func AddCharacteristic(uuid string) (string, error) {
	if uuid == "" {
		return "", fmt.Errorf("UUID cannot be empty")
	}
	return fmt.Sprintf("AT+QBLEGATTSCHAR=%s\r\n", uuid), nil
}

// FinishGATTServer 生成提交 GATT 服务定义的 AT 命令
func FinishGATTServer() string {
	return "AT+QBLEGATTSSRVDONE\r\n"
}

// SendNotify 生成发送 Notify 通知的 AT 命令
func SendNotify(handle string, value string) (string, error) {
	if handle <= "" {
		return "", fmt.Errorf("invalid handle: %s", handle)
	}
	if value == "" {
		return "", fmt.Errorf("value cannot be empty")
	}
	return fmt.Sprintf("AT+QBLEGATTSNTFY=0,%s,%s\r\n", handle, value), nil
}
