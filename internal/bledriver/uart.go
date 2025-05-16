package bledriver

import (
	"io"
	"log"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
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

// 传入：设备名、波特率
// 返回：Uart结构体指针
func NewUart(dev string, baud int, timeout int, lc logger.LoggingClient) (*Uart, error) {
	config := &serial.Config{
		Name:        dev,
		Baud:        baud,
		ReadTimeout: time.Duration(timeout) * time.Second, // 表示timeout s
	}

	var err error
	//尝试开启串口,成功返回(*serial.Port,nill),失败返回(nill,error)
	conn, err := serial.OpenPort(config)
	if err != nil || conn == nil {
		lc.Errorf("NewUart(): Exit - 开启 %s 串口失败 : %v", config.Name, err)
		return nil, err

	}
	//成功创建
	return &Uart{config: config, conn: conn, enable: true, portStatus: false}, nil
}

// 传入：读取串口缓冲取到rxbuf中
// 返回： error类或nil
func (dev *Uart) UartRead(maxbytes int, lc logger.LoggingClient) error {
	if !dev.enable {
		lc.Errorf("UartRead(): 串口 %s 未开启！！！", dev.config.Name)
		return nil
	}

	var buf []byte

	// serial包方法 一次读取的最大值为16byte
	// 分包读取
	readCount := (maxbytes / 16) + 1

	lc.Debugf("UartRead(): 串口 每次读值长度为: %v", readCount)

	if dev.portStatus {
		lc.Errorf("UartRead():  Exit - Device busy... Read request dropped for %s", dev.config.Name)
		return nil
	}

	dev.portStatus = true //？

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
				// 记录调试日志，提示成功读取到文件末尾
				lc.Debugf("UartRead(): %v - 完成读值!", err)
				// 跳出循环，结束读取
				break
			}
			// 对于其他类型的错误，记录错误日志
			lc.Errorf("UartRead(): Exit - Error = %v", err)

			// 将设备端口状态设置为不可用（false）
			dev.portStatus = false

			// 清空串口连接的缓冲区
			dev.conn.Flush()

			// 返回错误，终止函数执行
			return err
		}

		// 记录调试日志，显示本次读取的字节长度和具体数据内容
		lc.Debugf("UartRead(): 读取长度为 = %v, 值为 = %s", lens, b)

		// 将读取到的数据（b[:lens]）追加到临时缓冲区 buf 中
		// 使用切片 b[:lens] 确保只追加实际读取的字节
		buf = append(buf, b[:lens]...)

		// 将临时缓冲区 buf 的全部内容追加到设备的接收缓冲区 dev.rxbuf 中
		// 这里使用了 buf[:] 确保追加整个缓冲区内容
		dev.rxbuf = append(dev.rxbuf, buf[:]...)

		// 记录调试日志，显示设备接收缓冲区的当前内容
		lc.Debugf("UartRead(): dev.rxbuf = %s", dev.rxbuf)

		// 清空临时缓冲区 buf，为下一次读取做准备
		buf = nil
	}
	dev.portStatus = false
	dev.conn.Flush()
	lc.Debugf("UartRead(): Exit - Success")

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
func (dev *Uart) UartWrite(txbuf []byte, lc logger.LoggingClient) (int, error) {
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

	// 记录调试日志，显示成功写入的字节数
	lc.Debugf("UartWrite(): Number of bytes transmitted = %d\n", length)

	// 返回写入的字节数和错误（此时 err 为 nil）
	return length, err
}
