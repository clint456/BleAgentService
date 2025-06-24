package command

import (
	"fmt"
	"strings"
)

// 命令处理器协程
func commandProcessor() {
	for cmd := range inputChan {
		cmd = strings.TrimSpace(cmd)
		if cmd == "" {
			continue
		}

		fmt.Printf("[收到命令] %s\n", cmd)

		switch cmd {
		case "AT":
			writeResponse("\r\nOK\r\n")
		case "AT+STATUS":
			writeResponse("\r\n+STATUS:OK\r\n\r\nOK\r\n")
		case "AT+REBOOT":
			writeResponse("\r\nREBOOTING...\r\n\r\nOK\r\n")
		default:
			writeResponse("\r\nERROR\r\n")
		}
	}
}
