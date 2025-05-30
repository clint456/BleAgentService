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

type BleCommand string

const (
	// init
	ATRESET        BleCommand = "AT+QRST\r\n"
	ATVERSION      BleCommand = "AT+QVERSION\r\n"
	ATINIT_1       BleCommand = "AT+QBLEINIT=1\r\n"           //初始化BLE的中心设备
	ATINIT_2       BleCommand = "AT+QBLEINIT=2\r\n"           //作为外围设备初始化
	ATINIT_4       BleCommand = "AT+QBLEINIT=4\r\n"           //设备初始化为多连接
	ATADV          BleCommand = "AT+QBLEADVPARAM=150,150\r\n" //设置 BLE 广播参数
	ATGATTSSRV     BleCommand = "AT+QBLEGATTSSRV=fff1\r\n"
	ATGATTSCHAR    BleCommand = "AT+QBLEGATTSCHAR=fff2\r\n"
	ATGATTSSRVDONE BleCommand = "AT+QBLEGATTSSRVDONE\r\n"
	ATNAME         BleCommand = "AT+QBLENAME=QuecHCM111Z\r\n"
	ATADDR         BleCommand = "AT+QBLEADDR?\r\n"
	ATADVSTART     BleCommand = "AT+QBLEADVSTART\r\n"

	//query
	ATSTATE BleCommand = "AT+QBLESTAT\r\n" // 查询设备状态

	// send
	GATTSNTFY BleCommand = "AT+QBLEGATTSNTFY=0,fff2,"
)

type AtCommand struct {
	state   BleStatus
	at_uart *Uart
	lc      logger.LoggingClient
}

// 构造AtCommand的工厂函数
func NewAtCommand(at_uart *Uart, lc logger.LoggingClient) *AtCommand {
	state, _ := CheckAtState(at_uart) //检查当前Ble状态
	lc.Debug("当前BLE设备状态为: %v", state)
	return &AtCommand{
		state:   state,
		at_uart: at_uart,
		lc:      lc,
	}
}

// BLE初始化为外围连接设备(模式2)，并开启广播
func (a *AtCommand) BleInit_2() error {
	// if a.state == Uninitialized {
	// 	return fmt.Errorf("未初始化BLE模块")
	// }

	// TODO
	// 可以根据当前设备状态来决定前面几个命令需不要
	// 目前暂时就这样吧

	// 定义命令切片
	commands := []BleCommand{
		ATRESET,
		ATVERSION,
		ATINIT_2,
		ATADV,
		ATGATTSSRV,
		ATGATTSCHAR,
		ATGATTSSRVDONE,
		ATNAME,
		ATADDR,
		ATADVSTART,
	}

	var lastErr error

	// 使用循环发送命令
	for _, cmd := range commands {
		_, info, err := AtCommandSend(cmd, a.at_uart)
		a.lc.Debugf("✅发送At指令中 %v 的结果是:%v , error:%v ", cmd, info, err)
		if err != nil {
			lastErr = err // 保存最后一次的错误（如果有）
			// 可以选择在这里中断循环，或者继续执行其他命令
			a.lc.Errorf("❌发送At指令 %v 出现错误: %v ,输出为: %v", cmd, err, info)
			// 这里选择抛出错误并继续执行所有命令，保留最后一次的错误
		}
	}
	return lastErr
}

// 向Ble设备发送消息数据（string）
func (a *AtCommand) BleSendString(meg string) error {
	// 检查内容是否为空
	if meg == "" {
		return fmt.Errorf("发送内容不能为空")
	}

	// 拼接AT命令+消息，显式转换BleCommand为string并添加结尾
	command := string(GATTSNTFY) + meg + "\r\n"
	_, info, err := AtSendMesg(command, a.at_uart)
	if err != nil {
		a.lc.Errorf("❌BleSendString(): %v 失败: %v , 输出为: %v", command, err, info)
		return fmt.Errorf("发送mesg失败: %w", err)
	}
	a.lc.Debugf("✅ BleSendString(): %v 成功 :%v , 错误:%v ", command, info, err)

	return nil
}

// 获取ble模块当前状态
func CheckAtState(u *Uart) (BleStatus, error) {

	_, rxbuf, er := AtCommandSend(ATSTATE, u) //向Ble模块发送检查指令
	for _, status := range []BleStatus{Uninitialized, Initialized, Advertising, Connected, Disconnected} {
		if strings.Contains(rxbuf, string(status)) {
			return status, nil
		}
	}
	return Uninitialized, er // 默认值

}

// 向设备发送AT指令
// 返回：发送At指令长度、回显值、错误信息
func AtCommandSend(command BleCommand, u *Uart) (txlen int, rxbuf string, er error) {
	var err error
	u.rxbuf = nil // 先将接收缓存区置空
	// 写入状态查询AT指令 (使用切片发送)
	_txlen, err := u.UartWrite([]byte(command))
	if err != nil {
		return _txlen, fmt.Sprintln("fail"), fmt.Errorf("AtCommandSend(): AT指令写入串口失败 %v", err)
	}
	if err := u.UartRead(64); err != nil {
		return _txlen, fmt.Sprintln("fail"), fmt.Errorf("AtCommandSend(): AT 串口读值有错误 %v", err)
	}
	// 读值无错误
	_str := string(u.rxbuf)
	if !strings.Contains(_str, "OK") { // 蓝牙回显不OK
		return _txlen, fmt.Sprintln("fail"), fmt.Errorf("AtCommandSend(): 蓝牙回显错误: %v", _str)
	}
	// 清空 rxbuf 以准备下一次读取（可选）
	u.rxbuf = nil
	return _txlen, _str, err
}

// 向设备发送消息
// 返回：消息发送长度、回显值、错误信息
func AtSendMesg(mesg string, u *Uart) (txlen int, rxbuf string, er error) {
	var err error
	u.rxbuf = nil // 先将接收缓存区置空
	// 写入状态查询AT指令 (使用切片发送)
	_txlen, err := u.UartWrite([]byte(mesg))
	if err != nil {
		return _txlen, fmt.Sprintln("fail"), fmt.Errorf("AtSendMesg(): Message写入串口失败 %v", err)
	}
	if err := u.UartRead(64); err != nil {
		return _txlen, fmt.Sprintln("fail"), fmt.Errorf("AtSendMesg(): 串口读值有错误 %v", err)
	}

	// 读值无错误
	_str := string(u.rxbuf)
	if !strings.Contains(_str, "OK") { // 蓝牙回显不OK
		return _txlen, fmt.Sprintln("fail"), fmt.Errorf("AtSendMesg(): Message蓝牙回显错误: %v , 写入的数据为: %v", _str, mesg)
	}
	// 清空 rxbuf 以准备下一次读取（可选）
	u.rxbuf = nil
	return _txlen, _str, err
}
