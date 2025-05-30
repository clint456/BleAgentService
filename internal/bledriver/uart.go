package bledriver

import (
	"io"
	"log"
	"time"

	"github.com/tarm/serial"
)

// 串口控制模块结构体,
// config口初始化配置指针,
// conn:端口结构体指针,
// rxbuf:slice动态数组,
// portStatus：当autoevent被执行启用时，该值为true,
type Uart struct {
	Name       string
	config     *serial.Config
	conn       *serial.Port
	rxbuf      []byte
	enable     bool
	portStatus bool
}

type JsonMessage struct {
	APIVersion    string      `json:"apiVersion"`    // API 版本
	ReceivedTopic string      `json:"receivedTopic"` // 接收到的主题
	CorrelationID string      `json:"correlationID"` // 消息跟踪 ID
	RequestID     string      `json:"requestID"`     // 请求 ID
	ErrorCode     int         `json:"errorCode"`     // 错误码，0 表示成功
	Payload       interface{} `json:"payload"`       // 消息内容（动态类型）
	ContentType   string      `json:"contentType"`   // MIME 类型
}

// 传入：设备名、波特率
// 返回：Uart结构体指针
func NewUart(dev string, baud int, timeout int) (*Uart, error) {
	config := &serial.Config{
		Name:        dev,
		Baud:        baud,
		ReadTimeout: time.Duration(timeout) * time.Second, // 表示timeout s
	}

	var err error
	//尝试开启串口,成功返回(*serial.Port,nill),失败返回(nill,error)
	conn, err := serial.OpenPort(config)
	if err != nil || conn == nil {
		return nil, err
	}
	//成功创建
	return &Uart{config: config, conn: conn, enable: true, portStatus: false}, nil
}

func (dev *Uart) UartClose() {
	dev.conn.Close()
}

// 传入：读取串口缓冲取到rxbuf中
// 返回： error类或nil
func (dev *Uart) UartRead(maxbytes int) error {
	// if !dev.enable {
	// 	return nil
	// }
	var buf []byte
	// 分包读取
	readCount := (maxbytes / 16) + 1
	if dev.portStatus {
		return nil
	}
	dev.portStatus = true

	// 最多允许读取 128byte
	b := make([]byte, 128)

	// 循环读取数据，循环次数由 readCount 控制
	for i := 1; i <= readCount; i++ {
		// 从串口连接（dev.conn）读取数据到缓冲区 b，返回读取的字节数（lens）和可能的错误（err）
		lens, err := dev.conn.Read(b)

		// 检查读取过程中是否发生错误
		if err != nil {
			// 如果错误是文件结束（EOF），表示数据读取完成
			if err == io.EOF {
				// 跳出循环，结束读取
				break
			}
			// 将设备端口状态设置为不可用（false）
			dev.portStatus = false

			// 清空串口连接的缓冲区
			dev.conn.Flush()

			// 返回错误，终止函数执行
			return err
		}

		// 记录调试日志，显示本次读取的字节长度和具体数据内容
		// lc.Debugf("UartRead(): 读取长度为 = %v, 值为 = %s", lens, b)

		// 将读取到的数据（b[:lens]）追加到临时缓冲区 buf 中
		// 使用切片 b[:lens] 确保只追加实际读取的字节
		buf = append(buf, b[:lens]...)

		// 将临时缓冲区 buf 的全部内容追加到设备的接收缓冲区 dev.rxbuf 中
		// 这里使用了 buf[:] 确保追加整个缓冲区内容
		dev.rxbuf = append(dev.rxbuf, buf[:]...)

		// 清空临时缓冲区 buf，为下一次读取做准备
		buf = nil
	}
	dev.portStatus = false
	dev.conn.Flush()

	return nil
}

// UartWrite 是 Uart 结构体的方法，用于向串口设备写入数据
// 参数：
//   - txbuf: 要发送的字节切片（传输缓冲区）
//   - lc: 日志记录客户端，用于记录调试信息
//
// 返回值：
//   - int: 成功写入的字节数
//   - error: 写入过程中发生的错误（如果有）
func (dev *Uart) UartWrite(txbuf []byte) (int, error) {
	// 清空串口连接的输入和输出缓冲区，确保无残留数据干扰本次写入
	dev.conn.Flush()

	// 向串口连接（dev.conn）写入 txbuf 中的数据
	// 返回写入的字节数（length）和可能的错误（err）
	length, err := dev.conn.Write(txbuf)

	// 检查写入过程中是否发生错误
	if err != nil {
		// 如果发生错误，记录错误信息到标准日志
		log.Println(err)
		// 返回 0 表示无字节写入，并附带错误
		return 0, err
	}

	// 返回写入的字节数和错误（此时 err 为 nil）
	return length, err
}
