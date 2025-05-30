package bledriver

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/tarm/serial"
)

const (
	MTU        = 240
	Prefix     = "AT+QBLEGATTSNTFY=0,fff2,"
	Suffix     = "\r\n"
	HeaderSize = 4
	MaxPayload = MTU - len(Prefix) - len(Suffix) - HeaderSize
)

type Packet struct {
	Index   uint16
	Total   uint16
	Payload []byte
}

type ATUART struct {
	port  *serial.Port
	rxbuf []byte
}

func NewATUART(port *serial.Port) (*ATUART, error) {
	return &ATUART{port: port}, nil
}

func (u *ATUART) Write(data []byte) (int, error) {
	return u.port.Write(data)
}

func (u *ATUART) UartRead(size int) error {
	u.rxbuf = make([]byte, size)
	n, err := u.port.Read(u.rxbuf)
	if err != nil {
		return err
	}
	u.rxbuf = u.rxbuf[:n]
	return nil
}

func (u *ATUART) Close() error {
	return u.port.Close()
}

// splitIntoPackets 分包函数
func splitIntoPackets(data []byte) []Packet {
	var packets []Packet
	totalPackets := (len(data) + MaxPayload - 1) / MaxPayload
	for i := 0; i < len(data); i += MaxPayload {
		end := i + MaxPayload
		if end > len(data) {
			end = len(data)
		}
		packets = append(packets, Packet{
			Index:   uint16(i / MaxPayload),
			Total:   uint16(totalPackets),
			Payload: data[i:end],
		})
	}
	return packets
}

// SendJSONOverUART 发送 JSON 数据的主要函数
func SendJSONOverUART(port *serial.Port, jsonData []byte) error {
	dataBytes, err := json.Marshal(jsonData)
	if err != nil {
		return fmt.Errorf("JSON marshal error: %v", err)
	}

	packets := splitIntoPackets(dataBytes)

	uart, err := NewATUART(port)
	if err != nil {
		return fmt.Errorf("serial open error: %v", err)
	}

	for _, packet := range packets {
		packetData := make([]byte, len(Prefix)+HeaderSize+len(packet.Payload)+len(Suffix))
		copy(packetData, Prefix)
		binary.BigEndian.PutUint16(packetData[len(Prefix):], packet.Index)
		binary.BigEndian.PutUint16(packetData[len(Prefix)+2:], packet.Total)
		copy(packetData[len(Prefix)+HeaderSize:], packet.Payload)
		copy(packetData[len(Prefix)+HeaderSize+len(packet.Payload):], Suffix)

		_, err := uart.Write(packetData)
		if err != nil {
			log.Printf("Error sending packet %d: %v", packet.Index, err)
			continue
		}
		uart.rxbuf = nil
		if err := uart.UartRead(64); err != nil {
			log.Printf("Error reading response for packet %d: %v", packet.Index, err)
			continue
		}

		rxbuf := strings.TrimSpace(string(uart.rxbuf))
		if strings.Contains(rxbuf, "OK") {
			log.Printf("✅ Sent packet %d/%d OK", packet.Index+1, packet.Total)
		} else {
			log.Printf("❌ Invalid response for packet %d: %s", packet.Index, rxbuf)
		}

		uart.rxbuf = nil
		time.Sleep(1 * time.Millisecond)
	}

	return nil
}
