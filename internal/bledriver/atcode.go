package bledriver

import (
	"fmt"
	"strings"
	//"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
)

// BLE 设备状态
type BleStatus string

const (
	Uninitialized BleStatus = "NOINIT"
	Initialized   BleStatus = "INIT"
	Advertising   BleStatus = "ADVERTISING"
	Connected     BleStatus = "CONNECTED"
	Disconnected  BleStatus = "DISCONNECTED"
)

const (
	ATRESET        = "AT+QRST\r\n"
	ATVERSION      = "AT+QVERSION\r\n"
	ATINIT_1       = "AT+QBLEINIT=1\r\n"           //初始化BLE的中心设备
	ATINIT_2       = "AT+QBLEINIT=2\r\n"           //作为外围设备初始化
	ATINIT_4       = "AT+QBLEINIT=4\r\n"           //设备初始化为多连接
	ATADV          = "AT+QBLEADVPARAM=150,150\r\n" //设置 BLE 广播参数
	ATGATTSSRV     = "AT+QBLEGATTSSRV=fff1\r\n"
	ATGATTSCHAR    = "AT+QBLEGATTSCHAR=fff2\r\n"
	ATGATTSSRVDONE = "AT+QBLEGATTSSRVDONE\r\n"
	ATNAME         = "AT+QBLENAME=QuecHCM111Z\r\n"
	ATADDR         = "AT+QBLEADDR?\r\n"
	ATADVSTART     = "AT+QBLEADVSTART\r\n"
)

type AtCommand struct {
	state *BleStatus
}

func (a *AtCommand) AtCommandSend(code string, u *Uart, lc logger.LoggingClient) (string, error) {
	var err error

	// 写入状态查询AT指令 (使用切片发送)
	txlen, err := u.UartWrite([]byte(code), lc)
	if err != nil {
		lc.Errorf("AtCommandSend(): AT指令写入串口失败 %v", err)
		return fmt.Sprintln("fail"), fmt.Errorf("AtCommandSend(): AT指令写入串口失败 %v", err)
	}
	lc.Debugf("AtCommandSend(): AT指令 %v 已写入串口 length = %v",code, txlen)

	//TODO 可能在这里需要加上 300毫秒延时，视情况而定
	//
	//
	//
	// time.Sleep(300 * time.Millisecond)

	// 读取Ble模块回显值
	if err := u.UartRead(128, lc); err != nil {
		return fmt.Sprintln("fail"), fmt.Errorf("AtCommandSend(): AT 串口读值有错误 %v", err)
	}
	// 读值无错误
	_str := string(u.rxbuf)
	lc.Debugf("AtCommandSend(): 读取的回显值为： %v", _str)
	if !strings.Contains(_str, "OK") { // 蓝牙回显不OK
		return fmt.Sprintln("fail"), fmt.Errorf("AtCommandSend(): 蓝牙回显错误: %v", _str)
	}
	// 清空 rxbuf 以准备下一次读取（可选）
	u.rxbuf = nil

	return _str, err
}
