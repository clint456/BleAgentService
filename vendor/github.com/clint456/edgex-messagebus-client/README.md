# EdgeX MessageBus Client

[![Go Report Card](https://goreportcard.com/badge/github.com/clint456/edgex-messagebus-client)](https://goreportcard.com/report/github.com/clint456/edgex-messagebus-client)
[![GitHub License](https://img.shields.io/github/license/clint456/edgex-messagebus-client)](https://choosealicense.com/licenses/apache-2.0/)
[![Go Reference](https://pkg.go.dev/badge/github.com/clint456/edgex-messagebus-client.svg)](https://pkg.go.dev/github.com/clint456/edgex-messagebus-client)
[![GitHub release](https://img.shields.io/github/release/clint456/edgex-messagebus-client.svg)](https://github.com/clint456/edgex-messagebus-client/releases)

A high-level EdgeX Foundry MessageBus client library that provides a simple and intuitive API for MQTT and NATS message publishing and subscription operations.

一个高级的 EdgeX Foundry MessageBus 客户端库，提供简单易用的 API 来进行 MQTT 和 NATS 消息的发布和订阅操作。

## 🚀 特性

- ✅ **完整的 EdgeX MessageBus 支持** - 基于官方 `go-mod-messaging` 库
- ✅ **连接管理** - 自动连接、断开连接和重连机制
- ✅ **消息发布** - 支持多种数据类型的消息发布
- ✅ **消息订阅** - 支持主题订阅和自定义消息处理
- ✅ **请求-响应模式** - 支持同步请求-响应操作
- ✅ **二进制数据支持** - 支持发布和订阅二进制数据
- ✅ **线程安全** - 所有操作都是线程安全的
- ✅ **错误处理** - 完善的错误处理和日志记录
- ✅ **健康检查** - 提供客户端健康状态检查

## 📦 Installation | 安装

### Using Go Modules (Recommended)

```bash
go get github.com/clint456/edgex-messagebus-client
```

### Version Pinning

To use a specific version:

```bash
go get github.com/clint456/edgex-messagebus-client@v1.0.0
```

### Import in your Go code

```go
import messagebus "github.com/clint456/edgex-messagebus-client"
```

## 🛠️ Quick Start | 快速开始

### Basic Usage | 基本使用

```go
package main

import (
    "fmt"
    "log"
    "time"

    messagebus "github.com/clint456/edgex-messagebus-client"
    "github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
    "github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"
)

func main() {
    // 创建日志客户端
    lc := logger.NewClient("MyApp", "DEBUG")

    // 配置MessageBus客户端
    config := messagebus.Config{
        Host:     "localhost",
        Port:     1883,
        Protocol: "tcp",
        Type:     "mqtt",
        ClientID: "my-client",
        QoS:      1,
    }

    // 创建并连接客户端
    client, err := messagebus.NewClient(config, lc)
    if err != nil {
        log.Fatal(err)
    }

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
        fmt.Printf("收到消息: %s\n", string(message.Payload.([]byte)))
        return nil
    }
    client.SubscribeSingle("edgex/events/#", handler)

    // 等待消息
    time.Sleep(10 * time.Second)
}
```

## 📚 API 参考

### 配置结构

```go
type Config struct {
    Host     string  // MQTT Broker 主机地址
    Port     int     // MQTT Broker 端口
    Protocol string  // 协议 (tcp, ssl, ws, wss)
    Type     string  // 消息总线类型 (mqtt, nats)
    ClientID string  // 客户端 ID
    Username string  // 用户名 (可选)
    Password string  // 密码 (可选)
    QoS      int     // QoS 级别 (0, 1, 2)
}
```

### 主要方法

| 方法 | 描述 |
|------|------|
| `NewClient(config, logger)` | 创建新的客户端 |
| `Connect()` | 连接到 MessageBus |
| `Disconnect()` | 断开连接 |
| `IsConnected()` | 检查连接状态 |
| `Publish(topic, data)` | 发布消息 |
| `Subscribe(topics, handler)` | 订阅主题 |
| `Unsubscribe(topics...)` | 取消订阅 |
| `HealthCheck()` | 健康检查 |

### 高级方法

| 方法 | 描述 |
|------|------|
| `PublishWithCorrelationID()` | 使用指定 CorrelationID 发布 |
| `PublishBinaryData()` | 发布二进制数据 |
| `Request()` | 请求-响应操作 |
| `CreateMessageEnvelope()` | 创建消息信封 |
| `GetClientInfo()` | 获取客户端信息 |

## 🔧 高级用法

### 请求-响应模式

```go
// 创建请求消息
envelope, _ := client.CreateMessageEnvelope(requestData, "")

// 发送请求并等待响应
response, err := client.Request(
    envelope,
    "edgex/command/request",
    "edgex/command/response",
    5*time.Second,
)
```

### 二进制数据发布

```go
binaryData := []byte{0x01, 0x02, 0x03, 0x04}
client.PublishBinaryData("edgex/binary/data", binaryData)
```

### 错误监听

```go
go func() {
    for err := range client.GetErrorChannel() {
        log.Printf("MessageBus 错误: %v", err)
    }
}()
```

## 🔧 Advanced Usage | 高级用法

### Custom Message Handlers | 自定义消息处理器

```go
// Advanced message handler with error handling
handler := func(topic string, message types.MessageEnvelope) error {
    // Parse message payload
    var data map[string]interface{}
    if err := json.Unmarshal(message.Payload.([]byte), &data); err != nil {
        return fmt.Errorf("failed to parse message: %v", err)
    }

    // Process the message
    fmt.Printf("Processing message from %s: %+v\n", topic, data)

    // Return error if processing fails
    return nil
}
```

### Connection Management | 连接管理

```go
// Check connection status
if !client.IsConnected() {
    if err := client.Connect(); err != nil {
        log.Printf("Reconnection failed: %v", err)
    }
}

// Monitor connection health
go func() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        if err := client.HealthCheck(); err != nil {
            log.Printf("Health check failed: %v", err)
            // Implement reconnection logic here
        }
    }
}()
```

## 📊 Performance Considerations | 性能考虑

- Use appropriate buffer sizes for high-throughput scenarios
- Consider QoS levels based on your reliability requirements
- Implement proper error handling to avoid message loss
- Use connection pooling for multiple clients if needed
- Monitor memory usage with large message volumes

## 🔒 Security Best Practices | 安全最佳实践

- Always use TLS/SSL in production environments
- Implement proper authentication and authorization
- Validate and sanitize all incoming messages
- Use secure credential storage mechanisms
- Regularly update dependencies

## 🧪 Testing | 测试

Run the test suite:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run integration tests (requires EdgeX environment)
go test -tags=integration ./...
```

## 📈 Monitoring and Observability | 监控和可观测性

```go
// Monitor error channel
go func() {
    for err := range client.GetErrorChannel() {
        // Log error or send to monitoring system
        log.Printf("MessageBus error: %v", err)
        // metrics.IncrementErrorCounter()
    }
}()

// Get client statistics
info := client.GetClientInfo()
fmt.Printf("Client stats: %+v\n", info)
```

## 🔄 Migration Guide | 迁移指南

### From v0.x to v1.x

- Update import paths to use the new module structure
- Review configuration changes in the Config struct
- Update error handling patterns
- Check for deprecated methods

## 📚 Additional Resources | 其他资源

- [EdgeX Foundry Documentation](https://docs.edgexfoundry.org/)
- [Go Module Documentation](https://pkg.go.dev/github.com/clint456/edgex-messagebus-client)
- [MQTT Protocol Specification](https://mqtt.org/)
- [NATS Documentation](https://docs.nats.io/)

## 📄 License | 许可证

This project is licensed under the [Apache-2.0](LICENSE) License.

本项目采用 [Apache-2.0](LICENSE) 许可证。

## 🤝 Contributing | 贡献

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

欢迎贡献！请查看 [CONTRIBUTING.md](CONTRIBUTING.md) 了解详情。

## 📞 Support | 支持

- 📖 Documentation: [pkg.go.dev](https://pkg.go.dev/github.com/clint456/edgex-messagebus-client)
- 🐛 Bug Reports: [GitHub Issues](https://github.com/clint456/edgex-messagebus-client/issues)
- 💬 Discussions: [GitHub Discussions](https://github.com/clint456/edgex-messagebus-client/discussions)
- 📧 Email: Create an issue for support requests

## 🏆 Acknowledgments | 致谢

- EdgeX Foundry community for the excellent messaging framework
- Contributors who have helped improve this library
- Users who have provided feedback and bug reports

---

**Made with ❤️ for the EdgeX Foundry community**
