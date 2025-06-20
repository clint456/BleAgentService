package interfaces

import "time"

// SerialQueue 定义串口队列的通用接口
// 只暴露需要跨包调用的方法

type SerialQueue interface {
	SendCommand(command []byte, timeout time.Duration) (string, error)
	Close()
}
