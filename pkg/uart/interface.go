package uart

// 内部接口，不向外暴露
// SerialPortInterface 定义 SerialQueue 所依赖的串口接口
type SerialPortInterface interface {
	Write([]byte) (int, error)
	ReadLine() (string, error)
	Close() error
}
