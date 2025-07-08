package uart

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/tarm/serial"
)

// SerialPort 串口操作结构体，封装了串口连接及相关操作。
type SerialPort struct {
	port   *serial.Port         // 底层串口连接
	reader *bufio.Reader        // 带缓冲的读取器，用于读取串口数据
	mutex  sync.Mutex           // 互斥锁，确保线程安全
	logger logger.LoggingClient // 日志记录器，用于记录操作日志
}

// NewSerialPort 创建并初始化串口实例。
// 参数:
//   - cfg: 串口配置，包含端口名称、波特率、读取超时等信息
//   - logger: 日志记录器，用于记录串口操作日志
//
// 返回:
//   - *SerialPort: 新创建的串口实例
//   - error: 初始化过程中的错误（如果有）
func NewSerialPort(logger logger.LoggingClient) (*SerialPort, error) {
	// 创建串口配置
	c := &serial.Config{
		Name:        "/dev/ttyS3",      // 串口名称
		Baud:        115200,            // 波特率
		ReadTimeout: time.Duration(10), // 读取超时时间
	}

	// 打开串口连接
	port, err := serial.OpenPort(c)
	if err != nil {
		return nil, fmt.Errorf("打开串口失败: %w", err) // 包装错误并返回
	}

	// 初始化串口实例
	sp := &SerialPort{
		port:   port,                  // 保存串口连接
		reader: bufio.NewReader(port), // 创建带缓冲的读取器
		logger: logger,                // 设置日志记录器
	}

	return sp, nil
}

// Write 向串口写入数据，线程安全。
// 参数:
//   - data: 要写入的字节数组
//
// 返回:
//   - int: 成功写入的字节数
//   - error: 写入过程中的错误（如果有）
func (sp *SerialPort) Write(data []byte) (int, error) {
	sp.mutex.Lock()         // 加锁，确保线程安全
	defer sp.mutex.Unlock() // 解锁

	// 写入数据到串口
	n, err := sp.port.Write(data)
	if err != nil {
		sp.logger.Errorf("串口写入失败: %v", err) // 记录错误日志
		return n, err                       // 返回写入字节数和错误
	}
	// 记录写入成功的调试日志，显示写入字节数和数据内容
	sp.logger.Tracef("串口写入成功 %d 字节: %s", n, strings.TrimSpace(string(data)))
	return n, nil
}

// ReadLine 从串口读取一行数据，线程安全。
// 返回:
//   - string: 读取的一行数据（去除换行符和回车符）
//   - error: 读取过程中的错误（如果有）
func (sp *SerialPort) ReadLine() (string, error) {
	sp.mutex.Lock()         // 加锁，确保线程安全
	defer sp.mutex.Unlock() // 解锁

	// 读取直到遇到换行符
	line, err := sp.reader.ReadBytes('\n')
	if err != nil && err != io.EOF {
		sp.logger.Errorf("串口读取失败: %v", err) // 记录读取错误
		return "", err                      // 返回错误
	}
	// 去除行尾的换行符和回车符
	return strings.TrimRight(string(line), "\r\n"), nil
}

// Close 关闭串口连接，释放资源。
// 返回:
//   - error: 关闭过程中的错误（如果有）
func (sp *SerialPort) Close() error {
	sp.mutex.Lock()         // 加锁，确保线程安全
	defer sp.mutex.Unlock() // 解锁

	if sp.port == nil {
		return nil // 如果串口未打开，直接返回
	}

	// 关闭串口连接
	err := sp.port.Close()
	if err != nil {
		sp.logger.Errorf("关闭串口失败: %v", err) // 记录关闭错误
		return err                          // 返回错误
	}
	sp.logger.Info("串口已关闭") // 记录关闭成功的日志
	return nil
}
