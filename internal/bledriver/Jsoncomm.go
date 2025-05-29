package bledriver

import (
	"encoding/json"
	"log"

	"github.com/tarm/serial"
)

// ExtractJSONFromSerial 从串口读取数据流，提取完整的 JSON 包
// 参数:
//
//	ser: 已初始化的串口对象
//
// 返回:
//
//	[]map[string]interface{}: 提取到的 JSON 包列表
//	error: 读取或解析过程中的错误
func ExtractJSONFromSerial(ser *serial.Port) ([]map[string]interface{}, error) {
	var (
		buffer      string                   // 数据缓冲区
		braceCount  int                      // 跟踪花括号嵌套层级
		jsonPackets []map[string]interface{} // 存储提取的 JSON 包
	)

	// 读取缓冲区大小
	buf := make([]byte, 128)

	for {
		// 读取串口数据
		n, err := ser.Read(buf)
		if err != nil {
			return jsonPackets, err
		}
		if n == 0 {
			continue // 无数据，继续读取
		}

		// 将读取的数据追加到缓冲区
		data := string(buf[:n])
		buffer += data

		// 遍历读取的字符
		for _, char := range data {
			if char == '{' {
				braceCount++
			} else if char == '}' {
				braceCount--
			}

			// 当花括号匹配且缓冲区可能包含完整 JSON 时
			if braceCount == 0 && len(buffer) > 0 {
				// 查找 JSON 包的开头和结尾
				startIdx := -1
				endIdx := -1
				for i, c := range buffer {
					if c == '{' && startIdx == -1 {
						startIdx = i
					}
					if c == '}' {
						endIdx = i + 1
					}
				}

				// 提取可能的 JSON 字符串
				if startIdx != -1 && endIdx != -1 && endIdx > startIdx {
					jsonStr := buffer[startIdx:endIdx]
					var jsonObj map[string]interface{}
					// 验证 JSON 有效性
					if err := json.Unmarshal([]byte(jsonStr), &jsonObj); err == nil {
						jsonPackets = append(jsonPackets, jsonObj)
						// 清空已处理的缓冲区部分
						buffer = buffer[endIdx:]
					} else {
						log.Printf("无效 JSON: %v, 继续读取...", err)
					}
				}
			}
		}

		// 如果提取到 JSON 包，返回
		if len(jsonPackets) > 0 {
			return jsonPackets, nil
		}
	}
}
