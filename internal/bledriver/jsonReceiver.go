package bledriver

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/tarm/serial"
)

// ReceiveJSONOverUART 从串口接收并重组分包的JSON数据
func ReceiveJSONOverUART(port *serial.Port) ([]byte, error) {
	uart, err := NewATUART(port)
	if err != nil {
		return nil, fmt.Errorf("初始化UART失败: %v", err)
	}
	defer uart.Close()
	var expectedTotal uint16
	received := make(map[uint16]Packet)

	for {
		// 读取串口数据
		if err := uart.UartRead(MTU); err != nil {
			return nil, fmt.Errorf("读取串口数据失败: %v", err)
		}

		data := uart.rxbuf
		// 检查数据长度是否足以包含头部
		if len(data) < HeaderSize {
			log.Printf("收到数据包过短: %d 字节", len(data))
			continue
		}

		// 解析数据包
		packet := Packet{
			Index:   binary.BigEndian.Uint16(data[:2]),
			Total:   binary.BigEndian.Uint16(data[2:4]),
			Payload: data[HeaderSize:],
		}

		// 保存数据包
		received[packet.Index] = packet
		if expectedTotal == 0 {
			expectedTotal = packet.Total
		}

		// 检查是否接收到所有数据包
		if uint16(len(received)) == expectedTotal {
			// 按索引排序并重组数据
			var fullData []byte
			for i := uint16(0); i < expectedTotal; i++ {
				if p, ok := received[i]; ok {
					fullData = append(fullData, p.Payload...)
				} else {
					return nil, fmt.Errorf("缺失数据包 %d/%d", i, expectedTotal)
				}
			}

			// 验证JSON格式
			var jsonData interface{}
			if err := json.Unmarshal(fullData, &jsonData); err != nil {
				return nil, fmt.Errorf("JSON解析失败: %v", err)
			}

			return fullData, nil
		}

		// 清理缓冲区并等待下一包
		uart.rxbuf = nil
		time.Sleep(100 * time.Millisecond)
	}
}
