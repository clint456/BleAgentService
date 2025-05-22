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
	ATSTATE        = "AT+QBLESTAT\r\n"
)
s
type AtCommand struct {
	state   *BleStatus
	at_uart *Uart
}

// 工厂函数
func NewAtCommand(at_uart *Uart) *AtCommand{
	state := AtState()
	return &AtCommand{
		state: ,
		at
	}
}

func (a *AtCommand) BleInit(lc logger.LoggingClient) (string, error) {
	var info string
	var er error
	info, er = a.AtCommandSend(ATRESET, a.at_uart, lc)
	lc.Debugf(">>>>>>>>>>>>>>>>>>>>>>>> ATCommand: %v  的结果是:%v , error:%v ", ATRESET, info, er)

	info, er = a.AtCommandSend(ATVERSION, a.at_uart, lc)
	lc.Debugf(">>>>>>>>>>>>>>>>>>>>>>>> ATCommand: %v  的结果是:%v , error:%v ", ATVERSION, info, er)

	info, er = a.AtCommandSend(ATINIT_2, a.at_uart, lc)
	lc.Debugf(">>>>>>>>>>>>>>>>>>>>>>>> ATCommand: %v  的结果是:%v , error:%v ", ATINIT_2, info, er)

	info, er = a.AtCommandSend(ATADV, a.at_uart, lc)
	lc.Debugf(">>>>>>>>>>>>>>>>>>>>>>>> ATCommand: %v  的结果是:%v , error:%v ", ATADV, info, er)

	info, er = a.AtCommandSend(ATGATTSSRV, a.at_uart, lc)
	lc.Debugf(">>>>>>>>>>>>>>>>>>>>>>>> ATCommand: %v  的结果是:%v , error:%v ", ATGATTSSRV, info, er)

	info, er = a.AtCommandSend(ATGATTSCHAR, a.at_uart, lc)
	lc.Debugf(">>>>>>>>>>>>>>>>>>>>>>>> ATCommand: %v  的结果是:%v , error:%v ", ATGATTSCHAR, info, er)

	info, er = a.AtCommandSend(ATGATTSSRVDONE, a.at_uart, lc)
	lc.Debugf(">>>>>>>>>>>>>>>>>>>>>>>> ATCommand: %v  的结果是:%v , error:%v ", ATGATTSSRVDONE, info, er)

	info, er = a.AtCommandSend(ATNAME, a.at_uart, lc)
	lc.Debugf(">>>>>>>>>>>>>>>>>>>>>>>> ATCommand: %v  的结果是:%v , error:%v ", ATNAME, info, er)

	info, er = a.AtCommandSend(ATADDR, a.at_uart, lc)
	lc.Debugf(">>>>>>>>>>>>>>>>>>>>>>>> ATCommand: %v  的结果是:%v , error:%v ", ATADDR, info, er)

	info, er = a.AtCommandSend(ATADVSTART, a.at_uart, lc)
	lc.Debugf(">>>>>>>>>>>>>>>>>>>>>>>> ATCommand: %v  的结果是:%v , error:%v ", ATADVSTART, info, er)

	return info, er
}

// 获取ble模块当前状态
func (a *AtCommand) AtState(u *Uart, lc logger.LoggingClient) (BleStatus, error) {
	var info string
	var er error
	info, er = a.AtCommandSend(ATSTATE, u, lc)

	for _, status := range []BleStatus{Uninitialized, Initialized, Advertising, Connected, Disconnected} {
		if strings.Contains(info, string(status)) {

			return status, nil
		}
	}
	return Uninitialized, er // 默认值

}

// 向设备发送AT指令并返回回显值
func (a *AtCommand) AtCommandSend(code string, u *Uart, lc logger.LoggingClient) (string, error) {
	var err error

	// 写入状态查询AT指令 (使用切片发送)
	txlen, err := u.UartWrite([]byte(code), lc)
	if err != nil {
		lc.Errorf("AtCommandSend(): AT指令写入串口失败 %v", err)
		return fmt.Sprintln("fail"), fmt.Errorf("AtCommandSend(): AT指令写入串口失败 %v", err)
	}
	lc.Debugf("AtCommandSend(): AT指令 %v 已写入串口 length = %v", code, txlen)

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
