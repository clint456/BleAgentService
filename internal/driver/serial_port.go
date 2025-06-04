package driver

import (
	"bufio"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/tarm/serial"
)

type SerialPort struct {
	port   *serial.Port
	reader *bufio.Reader
	lock   sync.Mutex
	Debug  bool
}

func NewSerialPort(portName string, baud int, debug bool, timeout time.Duration) (*SerialPort, error) {
	cfg := &serial.Config{
		Name:        portName,
		Baud:        baud,
		ReadTimeout: timeout,
	}
	p, err := serial.OpenPort(cfg)
	if err != nil {
		return nil, err
	}
	return &SerialPort{
		port:   p,
		reader: bufio.NewReader(p),
		Debug:  debug,
	}, nil
}

func (s *SerialPort) Write(data []byte) (int, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	n, err := s.port.Write(data)
	if s.Debug {
		log.Printf("SerialPort.Write: wrote %d bytes, data: %s", n, data)
	}
	return n, err
}

func (s *SerialPort) ReadLine() ([]byte, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	line, err := s.reader.ReadBytes('\n')
	if err != nil && err != io.EOF {
		return nil, err
	}
	if s.Debug {
		log.Printf("SerialPort.ReadLine: %d bytes, line: %s", len(line), strings.TrimRight(string(line), "\r\n"))
	}
	return line, err
}

func (s *SerialPort) Close() error {
	return s.port.Close()
}
