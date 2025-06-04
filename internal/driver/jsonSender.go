package driver

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	MTU        = 247                                          // 蓝牙模块 MTU 限制为 64 字节
	BlePort    = "/dev/ttyS3"                                 // 替换为实际串口设备路径（如 COM3 或 /dev/ttyS0）
	BaudRate   = 115200                                       // 串口波特率，根据蓝牙模块配置调整
	Prefix     = "AT+QBLEGATTSNTFY=0,fff2,"                   // AT 指令前缀
	Suffix     = "\r\n"                                       // AT 指令后缀
	HeaderSize = 4                                            // 分包头部：2 字节索引 + 2 字节总包数
	MaxPayload = MTU - len(Prefix) - len(Suffix) - HeaderSize // 实际载荷大小：64 - 20 - 2 - 4 = 38 字节
)

// Packet 分包结构
type Packet struct {
	Index   uint16 // 分包索引
	Total   uint16 // 总分包数
	Payload []byte // 数据载荷
}

// splitIntoPackets 将数据分包
func splitIntoPackets(data []byte) []Packet {
	var packets []Packet
	totalPackets := (len(data) + MaxPayload - 1) / MaxPayload // 向上取整

	for i := 0; i < len(data); i += MaxPayload {
		end := i + MaxPayload
		if end > len(data) {
			end = len(data)
		}

		packet := Packet{
			Index:   uint16(i / MaxPayload),
			Total:   uint16(totalPackets),
			Payload: data[i:end],
		}
		packets = append(packets, packet)
	}

	return packets
}

// SendJSONOverUART 发送 JSON 数据的主要函数
func SendJSONOverUART(sq *SerialQueue, jsonData map[string]interface{}) error {
	tag := uuid.New().String()
	// 将 JSON 数据序列化为字节

	dataBytes, err := json.Marshal(jsonData)
	if err != nil {
		log.Fatalf("Error marshaling JSON: %v", err)
	}
	// 分包
	packets := splitIntoPackets(dataBytes)

	if err != nil {
		log.Fatalf("Error opening serial port: %v", err)
	}

	// 发送分包并验证回显
	for _, packet := range packets {
		// 构造分包数据：前缀 + 头部（索引 + 总包数） + 载荷 + 后缀
		packetData := make([]byte, len(Prefix)+HeaderSize+len(packet.Payload)+len(Suffix))
		copy(packetData, Prefix)
		binary.BigEndian.PutUint16(packetData[len(Prefix):], packet.Index)   // 2 字节索引
		binary.BigEndian.PutUint16(packetData[len(Prefix)+2:], packet.Total) // 2 字节总包数
		copy(packetData[len(Prefix)+HeaderSize:], packet.Payload)
		copy(packetData[len(Prefix)+HeaderSize+len(packet.Payload):], Suffix)
		// 通过串口发送
		response, err := sq.Send(packetData, time.Millisecond)
		if err != nil {
			log.Printf("❗️ Error sending packet %d: %v", packet.Index, err)
			continue
		}
		if response == "OK\n" {
			log.Printf("⚡ 数据包 %v 的子包发送 %v 成功", tag, packet.Index)
		}
		if response == "ERROR\n" {
			log.Printf("⛔️  数据包 %v 的发送子包 %v 失败", tag, packet.Index)
		}
		log.Printf("⬇️  Sent packet %d/%d, size: %d bytes\n", packet.Index+1, packet.Total, len(packetData))

		// 模拟蓝牙模块的发送间隔（根据实际模块调整）
		time.Sleep(1 * time.Millisecond)
	}

	log.Printf("✅️ All packets of Packet %v sent and verified.", tag)
	return nil
}

func readResponse(port *SerialPort, index uint16) (string, error) {
	var fullResponse string
	start := time.Now()
	timeout := 3 * time.Millisecond
	for {
		if time.Since(start) > timeout {
			return "", fmt.Errorf("❌ 数据发送回显超时")
		}
		rawLine, err := port.ReadLine()
		line := string(rawLine)
		if err != nil {
			if err == io.EOF {
				time.Sleep(time.Millisecond) // 小延时再读
				continue
			}
			return "", fmt.Errorf("❌ 数据发送回显读取失败: %w", err)
		}

		line = trimCRLF(line) // 注意这里传参

		if line == "" {
			continue // 跳过空行
		}

		fullResponse += line + "\n"

		// 检查是否是结尾状态
		if line == "OK" {
			log.Printf("⚡ 发送子包 %v 成功", index)
			return fullResponse, nil
		}
		if line == "ERROR" {
			log.Printf("⛔️ 发送子包 %v 失败", index)
			return fullResponse, fmt.Errorf("命令返回 ERROR")
		}
		if strings.HasPrefix(line, "+CME ERROR:") {
			return fullResponse, fmt.Errorf("模块错误: %s", line)
		}
	}
}
