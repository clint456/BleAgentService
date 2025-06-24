package uart

import (
	"device-ble/internal/interfaces"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
)

// SerialQueue 串口命令队列管理器，用于管理串口命令的发送和响应处理
type SerialQueue struct {
	serialPort SerialPortInterface           // 串口操作接口
	requestCh  chan interfaces.SerialRequest // 命令请求队列通道
	callback   func(string)                  // 异步消息回调函数
	stopCh     chan struct{}                 // 停止信号通道
	logger     logger.LoggingClient          // 日志记录器
	readerCh   chan string                   // 串口读取数据的通用管道
}

// NewSerialQueue 创建新的串口队列管理器并启动后台处理协程
// 参数:
//   - port: 串口操作接口，用于实际的串口读写
//   - logger: 日志记录器，用于记录操作日志
//   - cb: 异步消息回调函数，处理特定类型的串口响应
//
// 返回:
//   - *SerialQueue: 新创建的串口队列管理器实例
func NewSerialQueue(port SerialPortInterface, logger logger.LoggingClient, cb func(string)) *SerialQueue {
	q := &SerialQueue{
		serialPort: port,
		requestCh:  make(chan interfaces.SerialRequest, 10), // 初始化命令请求队列，容量为10
		stopCh:     make(chan struct{}),                     // 初始化停止信号通道
		callback:   cb,                                      // 设置异步消息回调
		logger:     logger,                                  // 设置日志记录器
	}

	go q.processRequests() // 启动后台协程，处理命令请求
	go q.startReaderLoop() // 启动后台协程，持续读取串口数据

	logger.Info("串口队列管理器已启动") // 记录管理器启动日志
	return q
}

// SendCommand 发送串口命令并等待响应
// 参数:
//   - command: 要发送的命令字节数组
//   - timeout: 等待响应的超时时间
//   - delay: 发送命令后等待读取的延迟时间
//
// 返回:
//   - string: 串口响应数据
//   - error: 执行过程中的错误（如果有）
func (q *SerialQueue) SendCommand(command []byte, timeout, delay time.Duration) (string, error) {
	if len(command) == 0 {
		return "", fmt.Errorf("命令不能为空") // 校验命令非空
	}

	responseCh := make(chan interfaces.SerialResponse, 1) // 创建响应通道
	req := interfaces.SerialRequest{
		Command:         command,    // 命令内容
		Timeout:         timeout,    // 超时时间
		DelayBeforeRead: delay,      // 读取前的延迟
		ResponseCh:      responseCh, // 响应通道
	}

	// 将请求发送到命令队列
	select {
	case q.requestCh <- req:
	case <-time.After(5 * time.Second):
		return "", fmt.Errorf("请求队列已满") // 队列满时返回错误
	}

	// 等待命令响应或超时
	select {
	case resp := <-responseCh:
		return resp.Data, resp.Error // 返回响应数据和错误
	case <-time.After(timeout + delay + time.Second):
		return "", fmt.Errorf("等待响应超时") // 超时返回错误
	}
}

// processRequests 后台协程，串行处理所有命令请求
// 持续从 requestCh 通道读取请求并执行
func (q *SerialQueue) processRequests() {
	for {
		select {
		case req := <-q.requestCh: // 接收新的命令请求
			resp := q.executeRequest(req) // 执行请求
			req.ResponseCh <- resp        // 将响应发送回请求者
		case <-q.stopCh: // 接收到停止信号
			return // 退出协程
		}
	}
}

// executeRequest 执行单个串口命令请求
// 参数:
//   - req: 包含命令、超时时间等信息的请求结构
//
// 返回:
//   - interfaces.SerialResponse: 包含响应数据和可能的错误
func (q *SerialQueue) executeRequest(req interfaces.SerialRequest) interfaces.SerialResponse {
	// 写入命令到串口
	if err := q.writeCommand(req.Command); err != nil {
		return interfaces.SerialResponse{Error: err} // 写入失败返回错误
	}

	// 如果指定了读取前的延迟，等待指定时间
	if req.DelayBeforeRead > 0 {
		time.Sleep(req.DelayBeforeRead)
	}

	var full strings.Builder           // 用于累积响应数据
	timeout := time.After(req.Timeout) // 设置超时定时器

	// 循环读取响应，直到收到终止响应或超时
	for {
		select {
		case <-timeout:
			return interfaces.SerialResponse{Data: "", Error: fmt.Errorf("读取响应超时")} // 超时返回错误
		case line := <-q.readerCh: // 从读取通道获取一行数据
			line = strings.TrimSpace(line) // 去除首尾空白
			if line == "" {
				continue // 忽略空行
			}
			q.logger.Debugf("响应: %s", line) // 记录调试日志
			full.WriteString(line + "\n")   // 累积响应数据

			// 判断是否为终止响应
			if q.isTerminal(line) {
				if strings.HasPrefix(line, "+CME ERROR:") {
					return interfaces.SerialResponse{Data: full.String(), Error: fmt.Errorf("模块错误: %s", line)} // 模块错误
				} else if line == "ERROR" {
					return interfaces.SerialResponse{Data: full.String(), Error: fmt.Errorf("命令执行失败")} // 命令失败
				} else if line == "OK" {
					return interfaces.SerialResponse{Data: full.String()} // 命令成功
				}
				return interfaces.SerialResponse{Data: full.String(), Error: fmt.Errorf("未知响应: %s", line)} // 未知响应
			}
		}
	}
}

// writeCommand 实际写入命令到串口
// 参数:
//   - cmd: 要发送的命令字节数组
//
// 返回:
//   - error: 写入过程中的错误（如果有）
func (q *SerialQueue) writeCommand(cmd []byte) error {
	_, err := q.serialPort.Write(cmd) // 写入命令
	if err != nil {
		q.logger.Errorf("串口写入失败: %v", err) // 记录错误日志
	}
	q.logger.Debugf("已发送命令: %s", strings.TrimSpace(string(cmd))) // 记录调试日志
	return err
}

// startReaderLoop 启动串口读取循环
// 创建常驻协程，持续读取串口数据并分发到 readerCh 或回调函数
func (q *SerialQueue) startReaderLoop() {
	q.readerCh = make(chan string, 100) // 初始化读取通道，缓冲容量100
	go func() {
		for {
			line, err := q.serialPort.ReadLine() // 读取一行串口数据
			if err != nil {
				if err == io.EOF { // 遇到EOF，短暂休眠后继续
					time.Sleep(10 * time.Millisecond)
					continue
				}
				q.logger.Errorf("串口读取错误: %v", err) // 记录读取错误
				continue
			}
			line = strings.TrimSpace(line) // 去除首尾空白
			if line == "" {
				continue // 忽略空行
			}
			// 忽略特定蓝牙模块的无用消息
			if strings.HasPrefix(line, "freqchip") {
				continue
			}
			// 判断是否为异步上报消息
			if strings.HasPrefix(line, "+COMMAND:") {
				if q.callback != nil {
					q.callback(line) // 调用回调函数处理异步消息
				}
			} else {
				q.readerCh <- line // 普通响应放入读取通道
			}
		}
	}()
}

// isTerminal 判断是否为终止响应（OK/ERROR/+CME ERROR）
// 参数:
//   - line: 串口响应的一行数据
//
// 返回:
//   - bool: 是否为终止响应
func (q *SerialQueue) isTerminal(line string) bool {
	return line == "OK" || line == "ERROR" || strings.HasPrefix(line, "+CME ERROR:")
}

// Close 关闭串口队列管理器
// 停止后台协程并清理资源
func (q *SerialQueue) Close() {
	close(q.stopCh)          // 发送停止信号
	q.logger.Info("串口队列已关闭") // 记录关闭日志
}
