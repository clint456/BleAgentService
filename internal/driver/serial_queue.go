package driver

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
)

// SerialRequest 串口请求
type SerialRequest struct {
	Command    []byte
	Timeout    time.Duration
	ResponseCh chan SerialResponse
}

// SerialResponse 串口响应
type SerialResponse struct {
	Data  string
	Error error
}

// SerialQueue 串口命令队列管理器
// 职责：管理串口命令的队列化执行，确保命令的顺序性和响应的正确性
type SerialQueue struct {
	serialPort *SerialPort
	requestCh  chan SerialRequest
	stopCh     chan struct{}
	logger     logger.LoggingClient
}

// NewSerialQueue 创建新的串口队列管理器
func NewSerialQueue(port *SerialPort, logger logger.LoggingClient) *SerialQueue {
	queue := &SerialQueue{
		serialPort: port,
		requestCh:  make(chan SerialRequest, 10), // 缓冲队列
		stopCh:     make(chan struct{}),
		logger:     logger,
	}

	go queue.processRequests()
	logger.Info("串口队列管理器已启动")
	return queue
}

// SendCommand 发送命令并等待响应
func (q *SerialQueue) SendCommand(command []byte, timeout time.Duration) (string, error) {
	if len(command) == 0 {
		return "", fmt.Errorf("命令不能为空")
	}

	responseCh := make(chan SerialResponse, 1)
	request := SerialRequest{
		Command:    command,
		Timeout:    timeout,
		ResponseCh: responseCh,
	}

	select {
	case q.requestCh <- request:
		// 请求已发送
	case <-time.After(5 * time.Second):
		return "", fmt.Errorf("请求队列已满，发送超时")
	}

	// 等待响应
	response := <-responseCh
	return response.Data, response.Error
}

// processRequests 处理串口请求队列
func (q *SerialQueue) processRequests() {
	for {
		select {
		case request := <-q.requestCh:
			response := q.executeRequest(request)
			request.ResponseCh <- response
		case <-q.stopCh:
			q.logger.Info("串口队列处理器已停止")
			return
		}
	}
}

// executeRequest 执行单个串口请求
func (q *SerialQueue) executeRequest(request SerialRequest) SerialResponse {
	// 写入命令
	if err := q.writeCommand(request.Command); err != nil {
		return SerialResponse{Error: fmt.Errorf("写入命令失败: %w", err)}
	}

	// 读取响应
	data, err := q.readResponse(request.Timeout)
	return SerialResponse{Data: data, Error: err}
}

// writeCommand 写入命令到串口
func (q *SerialQueue) writeCommand(command []byte) error {
	_, err := q.serialPort.Write(command)
	if err != nil {
		q.logger.Errorf("串口写入失败: %v", err)
		return err
	}

	q.logger.Debugf("命令已发送: %s", strings.TrimSpace(string(command)))
	return nil
}

// readResponse 读取串口响应
func (q *SerialQueue) readResponse(timeout time.Duration) (string, error) {
	var fullResponse strings.Builder
	timeoutCh := time.After(timeout)

	for {
		select {
		case <-timeoutCh:
			return "", fmt.Errorf("读取响应超时")
		default:
			line, err := q.readLine()
			if err != nil {
				if err == io.EOF {
					time.Sleep(10 * time.Millisecond)
					continue
				}
				return "", fmt.Errorf("读取失败: %w", err)
			}

			if line == "" {
				continue
			}

			fullResponse.WriteString(line + "\n")
			q.logger.Debugf("收到响应: %s", line)

			if q.isTerminalResponse(line) {
				return q.processTerminalResponse(fullResponse.String(), line)
			}
		}
	}
}

// readLine 读取一行数据
func (q *SerialQueue) readLine() (string, error) {
	rawLine, err := q.serialPort.ReadLine()
	if err != nil {
		return "", err
	}
	return strings.TrimRight(string(rawLine), "\r\n"), nil
}

// isTerminalResponse 检查是否为终端响应
func (q *SerialQueue) isTerminalResponse(line string) bool {
	return line == "OK" || line == "ERROR" || strings.HasPrefix(line, "+CME ERROR:")
}

// processTerminalResponse 处理终端响应
func (q *SerialQueue) processTerminalResponse(fullResponse, terminalLine string) (string, error) {
	switch {
	case terminalLine == "ERROR":
		return fullResponse, fmt.Errorf("设备返回错误")
	case strings.HasPrefix(terminalLine, "+CME ERROR:"):
		return fullResponse, fmt.Errorf("模块错误: %s", terminalLine)
	case terminalLine == "OK":
		return fullResponse, nil
	default:
		return fullResponse, fmt.Errorf("未知的终端响应: %s", terminalLine)
	}
}

// Close 关闭串口队列管理器
func (q *SerialQueue) Close() {
	close(q.stopCh)
	q.logger.Info("串口队列管理器已关闭")
}
