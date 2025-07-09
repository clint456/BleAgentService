package interfaces

// BLEController 定义了蓝牙低功耗（BLE）设备的统一操作接口。
// 该接口封装了蓝牙外围设备的初始化、命令发送以及队列管理等功能。
type BLEController interface {
	// Close 关闭蓝牙控制器，释放相关资源。
	//
	// 返回值:
	//   - error: 如果关闭过程中出现错误，返回非 nil 错误。
	Close() error

	// InitializeAsPeripheral 将蓝牙控制器初始化为外围设备模式。
	// 该方法通常用于设置蓝牙设备为外围设备，准备与中心设备进行交互。
	//
	// 返回值:
	//   - error: 如果初始化过程中出现错误，返回非 nil 错误。
	InitializeAsPeripheral() error

	// CustomInitializeBle 自定义初始化蓝牙设备。
	// 可以通过传入自定义的命令来初始化蓝牙设备。
	//
	// 参数:
	//   - cmd: 蓝牙设备初始化所需的命令字符串数组。
	//
	// 返回值:
	//   - error: 如果初始化过程中出现错误，返回非 nil 错误。
	CustomInitializeBle(cmd []string) error

	// SendSingle 发送单个命令到蓝牙设备。
	//
	// 参数:
	//   - cmd: 要发送的命令字符串(不超过247字节）。
	//
	// 返回值:
	//   - error: 如果发送过程中出现错误，返回非 nil 错误。
	SendSingle(cmd string) error

	// SendMulti 发送多个命令到蓝牙设备。
	//
	// 参数:
	//   - cmds: 要发送的命令字符串数组(单个数据元素不超过247字节）。
	//
	// 返回值:
	//   - error: 如果发送过程中出现错误，返回非 nil 错误。
	SendMulti(cmds []string) error

	// SendSingleWithResponse 发送单个命令并等待设备响应。
	//
	// 参数:
	//   - cmd: 要发送的命令字符串(不超过247字节）。
	//
	// 返回值:
	//   - res: 设备响应的字符串。
	//   - error: 如果发送或接收过程中出现错误，返回非 nil 错误。
	SendSingleWithResponse(cmd string) (res string, err error)

	// SendJSONOverBLE 分包发送 JSON 数据。
	// 该方法用于将 JSON 数据通过 BLE 发送，通常会将数据分割成适合的包进行发送。
	//
	// 参数:
	//   - Data: 要发送的 JSON 数据，可以是任何结构、任何大小的数据类型。
	//
	// 返回值:
	//   - error: 如果发送过程中出现错误，返回非 nil 错误。
	SendJSONOverBLE(Data interface{}) error

	// GetQueue 返回串口队列，具体类型由实现决定。
	// 该方法允许访问底层的串口队列管理接口，通常用于发送命令和管理数据流。
	//
	// 返回值:
	//   - SerialQueueInterface: 返回实现的串口队列接口。
	GetQueue() SerialQueueInterface
}
