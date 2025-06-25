package interfaces

import "time"

// BLEController 定义 BLE 控制器的通用接口
// 只暴露需要跨包调用的方法
// SerialQueue 定义串口队列的通用接口
// 只暴露需要跨包调用的方法
// SerialRequest 表示一个串口请求（命令 + 超时 + 响应通道）
type SerialRequest struct {
	Command         []byte              // 要发送的命令
	Timeout         time.Duration       // 响应超时时间
	DelayBeforeRead time.Duration       // 新增：写入命令后延迟读取时间
	ResponseCh      chan SerialResponse // 用于接收命令响应结果
}

// SerialResponse 表示串口返回的响应数据
type SerialResponse struct {
	Data  string // 响应内容（可能包含多行）
	Error error  // 错误信息（如超时、模块错误等）
}

type SerialQueue interface {
	SendCommand(command []byte, timeout, ReadDelay time.Duration) (string, error)
	Close()
}

type BLEController interface {
	InitializeAsPeripheral() error
	GetQueue() SerialQueue // 返回串口队列，具体类型由实现决定
}
