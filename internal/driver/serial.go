package driver

import (
	"bufio"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/tarm/serial"
)

// SerialPort 封装了串口操作，支持多 goroutine 并发访问时的线程安全。
type SerialPort struct {
	port    *serial.Port  // 底层串口对象
	mu      sync.Mutex    // 互斥锁，用于并发控制
	rxbuf   []byte        // 私有接收缓冲区
	timeout time.Duration // 读操作超时时间
}

// NewSerialPort 创建并打开串口，name 是串口名称，baud 是波特率，timeout 是读超时时间。
func NewSerialPort(name string, baud int, timeout time.Duration) (*SerialPort, error) {
	config := &serial.Config{
		Name:        name,
		Baud:        baud,
		ReadTimeout: timeout,
	}
	port, err := serial.OpenPort(config)
	if err != nil {
		log.Printf("Failed to open port %s: %v", name, err)
		return nil, err
	}
	return &SerialPort{
		port:    port,
		timeout: timeout,
	}, nil
}

// Write 向串口写入数据，返回写入的字节数和可能的错误。此方法会加锁。
func (sp *SerialPort) Write(cmd []byte) (int, error) {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	n, err := sp.port.Write(cmd)
	if err != nil {
		log.Printf("SerialPort.Write error: %v", err)
		return n, err
	}
	log.Printf("SerialPort.Write: wrote %d bytes, data: %x", n, cmd)
	return n, nil
}

// ReadLine 从串口读取一行数据（直到换行符），并返回该行字节。此方法会加锁。
func (sp *SerialPort) ReadLine() ([]byte, error) {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	reader := bufio.NewReader(sp.port)
	line, err := reader.ReadBytes('\n')
	if err != nil {
		log.Printf("SerialPort.ReadLine error: %v", err)
		return nil, err
	}
	log.Printf("SerialPort.ReadLine: %d bytes, line: %s", len(line), line)
	return line, nil
}

// ReadExact 持续读取指定长度的数据，直到达到长度或超时。超时后返回错误。
func (sp *SerialPort) ReadExact(size int) ([]byte, error) {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	buf := make([]byte, size)
	total := 0
	var deadline time.Time
	if sp.timeout > 0 {
		deadline = time.Now().Add(sp.timeout)
	}
	for total < size {
		if !deadline.IsZero() && time.Now().After(deadline) {
			err := fmt.Errorf("SerialPort.ReadExact timeout after %v", sp.timeout)
			log.Printf("%v", err)
			return nil, err
		}
		n, err := sp.port.Read(buf[total:])
		if err != nil {
			log.Printf("SerialPort.ReadExact error: %v", err)
			return nil, err
		}
		total += n
	}
	log.Printf("SerialPort.ReadExact: read %d bytes", total)
	return buf, nil
}

// UartRead 从串口读取指定长度的数据到私有缓冲区 rxbuf。此方法会加锁。
func (sp *SerialPort) UartRead(size int) error {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	buf := make([]byte, size)
	n, err := sp.port.Read(buf)
	if err != nil {
		log.Printf("SerialPort.UartRead error: %v", err)
		return err
	}
	sp.rxbuf = buf[:n] // 保存实际读取的数据
	log.Printf("SerialPort.UartRead: read %d bytes into rxbuf", n)
	return nil
}

// GetRxBuf 返回 rxbuf 缓冲区的副本，避免并发读写冲突。
func (sp *SerialPort) GetRxBuf() []byte {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	bufCopy := make([]byte, len(sp.rxbuf))
	copy(bufCopy, sp.rxbuf)
	return bufCopy
}

// Close 关闭串口。此方法会加锁，确保在读写过程中不会关闭串口。
func (sp *SerialPort) Close() error {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	if sp.port != nil {
		err := sp.port.Close()
		if err != nil {
			log.Printf("SerialPort.Close error: %v", err)
			return err
		}
		log.Printf("SerialPort.Close: port closed")
	}
	return nil
}
