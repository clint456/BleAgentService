package ble

import (
	"device-ble/internal/interfaces"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	MTU        = 247                                          // 蓝牙模块 MTU 限制为 64 字节
	Prefix     = "AT+QBLEGATTSNTFY=0,fff2,"                   // AT 指令前缀 (20字节）
	Suffix     = "\r\n"                                       // AT 指令后缀 (2字节)
	HeaderSize = 4                                            // 分包头部：2 字节索引 + 2 字节总包数
	MaxPayload = MTU - len(Prefix) - len(Suffix) - HeaderSize // 实际载荷大小：247 - 20 - 2 - 4 = 221 字节
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

// SendJSONOverBLE 发送 JSON 数据的主要函数。
func SendJSONOverBLE(sq interfaces.SerialQueue, jsonData interface{}) error {
	// 这里需要传入logger参数，建议后续重构接口
	// 目前先用fmt.Println模拟日志，后续可传入logger
	tag := uuid.New().String()
	dataBytes, err := json.Marshal(jsonData)
	if err != nil {
		return fmt.Errorf("JSON序列化失败: %v", err)
	}
	packets := splitIntoPackets(dataBytes)
	for _, packet := range packets {
		packetData := make([]byte, len(Prefix)+HeaderSize+len(packet.Payload)+len(Suffix))
		copy(packetData, Prefix)
		binary.BigEndian.PutUint16(packetData[len(Prefix):], packet.Index)
		binary.BigEndian.PutUint16(packetData[len(Prefix)+2:], packet.Total)
		copy(packetData[len(Prefix)+HeaderSize:], packet.Payload)
		copy(packetData[len(Prefix)+HeaderSize+len(packet.Payload):], Suffix)

		response, err := sq.SendCommand(packetData, 300*time.Millisecond, 1*time.Millisecond, 100*time.Millisecond)
		if strings.Contains(response, "OK") {
			fmt.Printf("⚡  数据包： %v  ⬇️ 子包：  %d/%d 发送成功, size：%d bytes\n", tag, packet.Index+1, packet.Total, len(packetData))
		} else if strings.Contains(response, "ERROR") {
			fmt.Printf("⛔️  数据包： %v  ⬇️ 子包：  %d/%d 发送失败\n", tag, packet.Index+1, packet.Total)
			return err
		} else {
			fmt.Printf("❗❓  数据包： %v  ⬇️ 子包：  %d/%d 未知回显：%v, error：%v\n", tag, packet.Index+1, packet.Total, response, err)
			return err
		}
	}
	fmt.Printf("✅️ All packets of Packet %v sent and verified.\n", tag)
	return nil
}
