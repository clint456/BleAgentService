package driver

import (
	"fmt"
	"io"
	"log"
	"strings"
	"time"
)

type SerialRequest struct {
	Cmd      []byte
	Timeout  time.Duration
	Response chan SerialResponse
}

type SerialResponse struct {
	Data string
	Err  error
}

type SerialQueue struct {
	port    *SerialPort
	reqChan chan SerialRequest
	quit    chan struct{}
}

func NewSerialQueue(port *SerialPort) *SerialQueue {
	q := &SerialQueue{
		port:    port,
		reqChan: make(chan SerialRequest),
		quit:    make(chan struct{}),
	}
	go q.loop()
	return q
}

func (q *SerialQueue) loop() {
	for {
		select {
		case req := <-q.reqChan:
			resp := q.handleRequest(req)
			req.Response <- resp
		case <-q.quit:
			return
		}
	}
}

func (q *SerialQueue) handleRequest(req SerialRequest) SerialResponse {
	_, err := q.port.Write(req.Cmd)
	if err != nil {
		return SerialResponse{Err: fmt.Errorf("写入失败: %w", err)}
	}

	var fullResponse string
	timeout := time.After(req.Timeout)

	for {
		select {
		case <-timeout:
			return SerialResponse{Err: fmt.Errorf("读取超时")}
		default:
			line, err := q.port.ReadLine()
			if err != nil {
				if err == io.EOF {
					time.Sleep(10 * time.Millisecond)
					continue
				}
				return SerialResponse{Err: fmt.Errorf("读取失败: %w", err)}
			}
			str := strings.TrimRight(string(line), "\r\n")
			if str == "" {
				continue
			}
			if q.port.Debug {
				log.Printf("🧾 收到: %q", str)
			}
			fullResponse += str + "\n"

			// 判定响应结束条件
			if str == "OK" || str == "ERROR" || strings.HasPrefix(str, "+CME ERROR:") {
				if str == "ERROR" {
					return SerialResponse{Data: fullResponse, Err: fmt.Errorf("设备返回 ERROR")}
				}
				if strings.HasPrefix(str, "+CME ERROR:") {
					return SerialResponse{Data: fullResponse, Err: fmt.Errorf("模块错误: %s", str)}
				}
				return SerialResponse{Data: fullResponse, Err: nil}
			}
		}
	}
}

func (q *SerialQueue) Send(cmd []byte, timeout time.Duration) (string, error) {
	respChan := make(chan SerialResponse)
	q.reqChan <- SerialRequest{
		Cmd:      cmd,
		Timeout:  timeout,
		Response: respChan,
	}
	resp := <-respChan
	return resp.Data, resp.Err
}

func (q *SerialQueue) Close() {
	close(q.quit)
}
