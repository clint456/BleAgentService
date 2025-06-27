package uart

import (
	"device-ble/internal/interfaces"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
)

// SerialQueue 串口命令队列管理器，用于管理串口命令的发送和响应处理。
type SerialQueue struct {
	serialPort      interfaces.SerialPortInterface // 串口操作接口
	requestCh       chan interfaces.SerialRequest  // 命令请求队列通道
	pendingRequests []interfaces.SerialRequest     // 待处理请求，按顺序存储
	commandCallback func(string)                   // 异步命令消息回调函数
	upAgentCallback func(string)                   // 异步透明代理回调函数
	stopCh          chan struct{}                  // 停止信号通道
	logger          logger.LoggingClient           // 日志记录器
	readerCh        chan string                    // 串口读取数据的通用管道
}

// NewSerialQueue 创建新的串口队列管理器并启动后台处理协程。
func NewSerialQueue(port interfaces.SerialPortInterface, logger logger.LoggingClient, ccb, uacb func(string), queueSize int) *SerialQueue {
	if queueSize <= 0 {
		queueSize = 10 // 默认容量 10
	}
	q := &SerialQueue{
		serialPort:      port,
		requestCh:       make(chan interfaces.SerialRequest, queueSize),
		pendingRequests: make([]interfaces.SerialRequest, 0),
		commandCallback: ccb,
		upAgentCallback: uacb,
		stopCh:          make(chan struct{}),
		logger:          logger,
	}
	go q.processRequests()
	go q.startReaderLoop()
	q.logger.Infof("串口队列管理器已启动，请求队列容量: %d", queueSize)
	return q
}

// SendCommand 发送串口命令并等待响应。
// 参数:
//   - command: 要发送到串口设备的命令字节数组，不能为空。
//   - timeout: 等待设备响应的最大时间，单位为纳秒。
//   - readDelay: 发送命令后等待设备处理的延迟时间，单位为纳秒。
//   - queueTimeout: 尝试将请求放入队列的最大等待时间，单位为纳秒。
//
// 返回值:
//   - string: 设备返回的响应数据（例如 "OK" 或 "ERROR"）。
//   - error: 如果发生错误（如命令为空、队列满、响应超时），返回非 nil 错误。
//
// 错误:
//   - "命令不能为空": 如果 command 参数为空。
//   - "请求队列已满": 如果在 queueTimeout 时间内无法将请求放入队列（最多重试 3 次）。
//   - "等待响应超时": 如果在 readDelay + timeout 时间内未收到设备响应。
func (q *SerialQueue) SendCommand(command []byte, timeout, readDelay, queueTimeout time.Duration) (string, error) {
	if len(command) == 0 {
		return "", fmt.Errorf("命令不能为空")
	}
	responseCh := make(chan interfaces.SerialResponse, 1) // 容量为 1，确保单一响应
	req := interfaces.SerialRequest{
		Command:         command,
		Timeout:         timeout,
		DelayBeforeRead: readDelay,
		ResponseCh:      responseCh,
		Timestamp:       time.Now(), // 用于超时清理
	}
	for retries := 0; retries < 3; retries++ {
		select {
		case q.requestCh <- req:
			q.logger.Debugf("请求发送成功，命令: %s, 重试次数: %d", string(command), retries)
			goto WaitResponse
		case <-time.After(queueTimeout):
			q.logger.Warnf("请求队列已满，当前长度: %d/%d, 重试次数: %d", len(q.requestCh), cap(q.requestCh), retries)
			if retries == 2 {
				return "", fmt.Errorf("请求队列已满，当前长度: %d/%d, 重试 3 次失败", len(q.requestCh), cap(q.requestCh))
			}
		}
	}
WaitResponse:
	select {
	case resp := <-responseCh:
		return resp.Data, resp.Error
	case <-time.After(readDelay + timeout):
		return "", fmt.Errorf("等待响应超时，命令: %s", string(command))
	}
}

// processRequests 后台协程，串行处理所有命令请求。
// 从 requestCh 读取请求，写入串口命令，并将成功写入的请求加入 pendingRequests 等待响应。
func (q *SerialQueue) processRequests() {
	for {
		select {
		case <-q.stopCh:
			q.logger.Debugf("停止处理请求协程")
			return
		case req := <-q.requestCh:
			if err := q.writeCommand(req.Command); err != nil {
				// 写入失败，直接发送错误响应，不加入 pendingRequests
				resp := interfaces.SerialResponse{Data: "", Error: fmt.Errorf("写入命令失败: %v", err)}
				select {
				case req.ResponseCh <- resp:
					q.logger.Debugf("响应发送成功（写入错误），命令: %s", string(req.Command))
				default:
					q.logger.Warnf("响应通道已满，命令: %s", string(req.Command))
				}
				continue // 处理下一个请求
			}
			// 写入成功，加入 pendingRequests
			q.pendingRequests = append(q.pendingRequests, req)
			q.logger.Debugf("请求加入待处理列表，命令: %s, 当前待处理数: %d", string(req.Command), len(q.pendingRequests))
		}
	}
}

// writeCommand 实际写入命令到串口。
func (q *SerialQueue) writeCommand(cmd []byte) error {
	_, err := q.serialPort.Write(cmd)
	if err != nil {
		q.logger.Errorf("串口写入失败: %v", err)
	}
	q.logger.Debugf("已发送命令: %s", string(cmd))
	return err
}

// startReaderLoop 启动串口读取循环。
// 持续读取串口数据，处理终止响应并匹配到 pendingRequests 的最早请求。
func (q *SerialQueue) startReaderLoop() {
	q.readerCh = make(chan string, 100)
	go func() {
		for {
			select {
			case <-q.stopCh:
				q.logger.Debugf("停止读取协程")
				return
			default:
				// 清理超时的 pendingRequests
				now := time.Now()
				for len(q.pendingRequests) > 0 && now.Sub(q.pendingRequests[0].Timestamp) > q.pendingRequests[0].Timeout+q.pendingRequests[0].DelayBeforeRead {
					req := q.pendingRequests[0]
					resp := interfaces.SerialResponse{Data: "", Error: fmt.Errorf("请求超时，未收到响应，命令: %s", string(req.Command))}
					select {
					case req.ResponseCh <- resp:
						q.logger.Warnf("请求超时，命令: %s, 已移除", string(req.Command))
					default:
						q.logger.Warnf("响应通道已满，超时请求，命令: %s", string(req.Command))
					}
					q.pendingRequests = q.pendingRequests[1:]
				}

				line, err := q.serialPort.ReadLine()
				if err != nil {
					if err == io.EOF {
						time.Sleep(1 * time.Millisecond)
						continue
					}
					q.logger.Errorf("串口读取错误: %v", err)
					continue
				}
				line = strings.Trim(line, "\r\n")
				if line == "" || strings.Contains(line, "freqchip") {
					continue
				}
				q.logger.Debugf("收到串口数据: %s", line)
				if strings.Contains(line, "+COMMAND:") {
					lines := strings.Split(line, "+COMMAND:")
					for _, part := range lines {
						if part == "" {
							continue
						}
						if q.commandCallback != nil {
							go q.commandCallback(part)
						}
					}
					continue
				}
				if q.upAgentCallback != nil {
					go q.upAgentCallback(line)
				}
				// 按顺序匹配最早的待处理请求
				if len(q.pendingRequests) > 0 && q.isTerminal(line) {
					req := q.pendingRequests[0]
					// 应用 DelayBeforeRead
					if req.DelayBeforeRead > 0 {
						time.Sleep(req.DelayBeforeRead)
						q.logger.Debugf("应用读取延迟: %v, 命令: %s", req.DelayBeforeRead, string(req.Command))
					}
					resp := interfaces.SerialResponse{Data: line}
					if strings.Contains(line, "ERROR") {
						resp.Error = fmt.Errorf("命令执行失败: %s", line)
					}
					select {
					case req.ResponseCh <- resp:
						q.logger.Debugf("响应发送成功，命令: %s, 数据: %s", string(req.Command), line)
						q.pendingRequests = q.pendingRequests[1:] // 移除已处理请求
					default:
						q.logger.Warnf("响应通道已满，命令: %s", string(req.Command))
						q.pendingRequests = q.pendingRequests[1:]
					}
				}
			}
		}
	}()
}

// isTerminal 判断是否为终止响应（OK/ERROR/+QBLEGATTSNTFY/其他）。
func (q *SerialQueue) isTerminal(line string) bool {
	return strings.Contains(line, "OK") ||
		strings.Contains(line, "ERROR") ||
		strings.Contains(line, "+QBLEGATTSNTFY") ||
		strings.Contains(line, "+CME ERROR")
}

// Close 关闭串口队列管理器，停止后台协程并清理资源。
func (q *SerialQueue) Close() error {
	q.logger.Infof("关闭串口队列管理器")
	close(q.stopCh)
	if err := q.serialPort.Close(); err != nil {
		q.logger.Errorf("关闭串口失败: %v", err)
		return fmt.Errorf("关闭串口失败: %v", err)
	}
	// 清空待处理请求
	for _, req := range q.pendingRequests {
		select {
		case req.ResponseCh <- interfaces.SerialResponse{Data: "", Error: fmt.Errorf("串口已关闭")}:
			q.logger.Debugf("通知请求关闭，命令: %s", string(req.Command))
		default:
			q.logger.Warnf("响应通道已满或关闭，命令: %s", string(req.Command))
		}
	}
	q.pendingRequests = nil
	close(q.readerCh)
	q.logger.Infof("串口队列管理器已关闭")
	return nil
}

func (q *SerialQueue) GetPort() interfaces.SerialPortInterface {
	return q.serialPort
}
