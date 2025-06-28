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
	Timestamp       time.Time           // 用于超时清理
}

// SerialResponse 表示串口返回的响应数据
type SerialResponse struct {
	Data  string // 响应内容（可能包含多行）
	Error error  // 错误信息（如超时、模块错误等）
}

// SerialPortInterface 定义 SerialQueue 所依赖的串口接口
type SerialPortInterface interface {
	Write([]byte) (int, error)
	ReadLine() (string, error)
	Close() error
}

type SerialQueueInterface interface {
	// SendCommand 发送串口命令并等待设备响应。支持并发发送
	//
	// 参数:
	//   - command: 要发送到串口设备的命令字节数组，不能为空。
	//   - timeout: 等待设备响应的最大时间。
	//   - readDelay: 发送命令后等待设备处理的延迟时间。
	//   - queueTimeout: 尝试将请求放入队列的最大等待时间。
	//
	// 返回值:
	//   - string: 设备返回的响应数据（例如 "OK" 或 "ERROR"）。
	//   - error: 如果发生错误（如命令为空、队列满、响应超时），返回非 nil 错误。
	//
	// 错误:
	//   - "命令不能为空": 如果 command 参数为空。
	//   - "请求队列已满": 如果在 queueTimeout 时间内无法将请求放入队列（最多重试 3 次）。
	//   - "等待响应超时": 如果在 readDelay + timeout 时间内未收到设备响应。
	SendCommand(command []byte, timeout, readDelay, queueTimeout time.Duration) (string, error)
	GetPort() SerialPortInterface
	Close() error
}

type BLEController interface {
	Close() error
	InitializeAsPeripheral() error
	CustomInitializeBle(cmd []string) error
	SendSingle(cmd string) error
	SendMulti(cmds []string) error
	GetQueue() SerialQueueInterface // 返回串口队列，具体类型由实现决定
}
