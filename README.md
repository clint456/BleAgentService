# BLE Agent Service

[![Go Report Card](https://goreportcard.com/badge/github.com/clint456/BleAgentService)](https://goreportcard.com/report/github.com/clint456/BleAgentService) [![GitHub License](https://img.shields.io/github/license/clint456/BleAgentService)](https://choosealicense.com/licenses/apache-2.0/) ![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/clint456/BleAgentService)

## 概述

BLE Agent Service 是一个基于 EdgeX Foundry 的设备服务，专门用于蓝牙低功耗（BLE）设备的连接和管理。该服务集成了串口通信、MQTT 消息传输和 EdgeX MessageBus 功能，为 IoT 设备提供完整的连接解决方案。

### 🎯 主要功能

- **🔵 BLE 设备管理** - 支持 BLE 外围设备初始化和广播
- **📡 串口通信** - 高性能串口数据传输和队列管理
- **🌐 MQTT 集成** - 完整的 MQTT 客户端和消息转发功能
- **🚀 EdgeX MessageBus** - 原生 EdgeX 消息总线支持
- **⚙️ 设备配置** - 灵活的设备配置和管理
- **🔄 数据转发** - 自动数据转发和消息路由
- **📊 实时监控** - 设备状态监控和健康检查

### 🏗️ 系统架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   BLE 设备      │◄──►│  串口通信模块    │◄──►│  BLE 控制器     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  MQTT Broker    │◄──►│   MQTT 客户端   │◄──►│ EdgeX MessageBus│
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌─────────────────┐
                       │  EdgeX Core     │
                       │   Services      │
                       └─────────────────┘
```

## 🚀 核心组件

### 1. BLE 控制器 (BleController)
- **AT 命令支持** - 完整的 BLE AT 命令集
- **外围设备模式** - 自动初始化为 BLE 外围设备
- **广播管理** - 自动启动和管理 BLE 广播
- **GATT 服务** - 支持 GATT 服务和特征值配置

### 2. 串口通信 (SerialPort & SerialQueue)
- **高性能串口** - 基于 `github.com/tarm/serial` 的串口通信
- **线程安全** - 使用 mutex 保护并发访问
- **队列管理** - 串口数据队列和缓冲管理
- **超时控制** - 可配置的读写超时机制

### 3. MQTT 客户端
- **双客户端架构** - 监听客户端 + 转发客户端
- **自动重连** - 连接断开时自动重连
- **消息转发** - 自动将接收到的消息转发到 MessageBus
- **认证支持** - 支持用户名/密码认证

### 4. EdgeX MessageBus 客户端
- **完整 API** - 发布、订阅、请求-响应模式
- **多数据类型** - 支持 JSON、字符串、二进制数据
- **线程安全** - 所有操作都是线程安全的
- **错误处理** - 完善的错误处理和重试机制
- **健康检查** - 实时监控客户端状态

## 📋 系统要求

- **操作系统**: Linux (推荐 Ubuntu 18.04+)
- **Go 版本**: Go 1.21 或更高版本
- **硬件**: 支持串口的设备 (如 Raspberry Pi)
- **依赖服务**: EdgeX Core Services, MQTT Broker

## ⚙️ 配置说明

### 主配置文件 (`cmd/res/configuration.yaml`)

```yaml
MQTTBrokerInfo:
  Schema: "tcp"
  Host: "192.168.8.196"
  Port: 1883
  Qos: 0
  KeepAlive: 3600
  ClientId: "device-ble-agent"
  AuthMode: "none"
  IncomingTopic: "edgex/events/#"

  Writable:
    ResponseFetchInterval: 500 # milliseconds
```

### 设备配置 (`cmd/res/devices/devices.yaml`)

```yaml
deviceList:
  - name: "Uart-Ble-Device"
    profileName: "Uart-Ble-Device"
    description: "Example of Device UART"
    protocols:
      UART:
        deviceLocation: "/dev/ttyS3"
        baudRate: 115200
```

## 🛠️ 安装和部署

### 1. 克隆项目

```bash
git clone https://github.com/clint456/BleAgentService.git
cd BleAgentService
```

### 2. 安装依赖

```bash
go mod tidy
```

### 3. 编译项目

```bash
make build
# 或者
go build -o ble-agent-service ./cmd
```

### 4. 配置设备

编辑 `cmd/res/configuration.yaml` 文件，配置 MQTT Broker 信息和串口设备路径。

### 5. 启动服务

```bash
./ble-agent-service
```

## 📖 使用指南

### 1. EdgeX CLI 操作

```bash
# 列出设备命令
edgex-cli command list -d Uart-Ble-Device

# 读取设备数据
edgex-cli command get -d Uart-Ble-Device -c String

# 写入设备数据
edgex-cli command set -d Uart-Ble-Device -c String -v "Hello BLE"
```

### 2. RESTful API 操作

```bash
# 获取设备信息
curl -X 'GET' http://localhost:59882/api/v2/device/name/Uart-Ble-Device | json_pp

# 读取设备数据
curl -X 'GET' http://localhost:59882/api/v2/device/name/Uart-Ble-Device/String

# 写入设备数据
curl -X 'PUT' http://localhost:59882/api/v2/device/name/Uart-Ble-Device/String \
     -H 'Content-Type: application/json' \
     -d '{"String":"Hello BLE Device"}'
```

### 3. MessageBus 客户端使用

```go
// 创建 MessageBus 客户端
config := &driver.ServiceConfig{...}
lc := logger.NewClient("MyService", "DEBUG")
client, err := driver.NewMessageBusClient(config, lc)

// 连接到 MessageBus
if err := client.Connect(); err != nil {
    log.Fatal(err)
}
defer client.Disconnect()

// 发布消息
data := map[string]interface{}{
    "deviceName": "sensor01",
    "reading":    25.6,
    "timestamp":  time.Now(),
}
client.Publish("edgex/events/device/sensor01", data)

// 订阅消息
handler := func(topic string, message types.MessageEnvelope) error {
    fmt.Printf("收到消息: %s\n", string(message.Payload))
    return nil
}
client.SubscribeSingle("edgex/events/#", handler)
```

## 🔧 开发指南

### 项目结构

```
BleAgentService/
├── cmd/                    # 主程序入口
│   ├── main.go            # 主程序
│   └── res/               # 配置文件
├── internal/driver/       # 核心驱动代码
│   ├── driver.go          # 主驱动
│   ├── bleController.go   # BLE 控制器
│   ├── serial_port.go     # 串口通信
│   ├── mqttClient.go      # MQTT 客户端
│   ├── messageBusClient.go # MessageBus 客户端
│   └── config.go          # 配置管理
├── examples/              # 示例代码
├── docs/                  # 文档
├── go.mod                 # Go 模块
└── README.md              # 项目说明
```

### 添加新功能

1. **扩展 BLE 命令** - 在 `bleCommand.go` 中添加新的 AT 命令
2. **自定义消息处理** - 在 `mqttIncomingListener.go` 中修改消息处理逻辑
3. **新增设备资源** - 在 `cmd/res/profiles/generic.yaml` 中添加设备资源
4. **配置新参数** - 在 `config.go` 中添加配置结构

### 调试技巧

- 启用 DEBUG 日志级别查看详细信息
- 使用串口调试工具监控 AT 命令交互
- 通过 MQTT 客户端工具测试消息传输
- 使用 EdgeX CLI 工具进行设备操作测试

## 📚 API 参考

### MessageBus 客户端 API

| 方法 | 描述 | 示例 |
|------|------|------|
| `Connect()` | 连接到 MessageBus | `client.Connect()` |
| `Publish()` | 发布消息 | `client.Publish(topic, data)` |
| `Subscribe()` | 订阅主题 | `client.Subscribe(topics, handler)` |
| `Request()` | 请求-响应 | `client.Request(msg, reqTopic, respTopic, timeout)` |
| `HealthCheck()` | 健康检查 | `client.HealthCheck()` |

### BLE 控制器 API

| 方法 | 描述 | AT 命令 |
|------|------|---------|
| `InitAsPeripheral()` | 初始化外围设备 | `AT+QBLEINIT=2` |
| `sendCommand()` | 发送 AT 命令 | 自定义命令 |

## 🔍 故障排除

### 常见问题

1. **串口连接失败**
   - 检查设备路径 `/dev/ttyS3` 是否存在
   - 确认波特率设置正确 (115200)
   - 检查串口权限

2. **MQTT 连接失败**
   - 验证 MQTT Broker 地址和端口
   - 检查网络连接
   - 确认认证信息

3. **BLE 初始化失败**
   - 检查 AT 命令响应
   - 确认硬件模块正常工作
   - 查看串口通信日志

### 日志分析

```bash
# 启用 DEBUG 日志
export EDGEX_LOGGING_LEVEL=DEBUG

# 查看服务日志
journalctl -u ble-agent-service -f
```

## 🤝 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 📄 许可证

本项目采用 [Apache-2.0](LICENSE) 许可证。

## 🙏 致谢

- [EdgeX Foundry](https://www.edgexfoundry.org/) - 提供优秀的边缘计算框架
- [Jiangxing Intelligence](https://www.jiangxingai.com) - 原始 UART 设备服务贡献者
- HCL Technologies(EPL Team) - 项目贡献者

## 📞 联系方式

- 项目地址: [https://github.com/clint456/BleAgentService](https://github.com/clint456/BleAgentService)
- 问题反馈: [Issues](https://github.com/clint456/BleAgentService/issues)
- 文档: [docs/](docs/)

---

**注意**: 本服务仅在 Linux 系统上运行，推荐在 Raspberry Pi 等嵌入式设备上使用。


