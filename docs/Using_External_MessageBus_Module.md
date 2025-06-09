# 使用独立的 EdgeX MessageBus 客户端模块

## 🎯 概述

我们已经将 MessageBus 客户端功能提取为一个独立的 Go 模块，这样您就可以在任何项目中重用这个功能，而不需要将代码复制到每个项目中。

## 📦 模块信息

- **模块名称**: `github.com/clint456/edgex-messagebus-client`
- **版本**: v0.1.0
- **位置**: `/home/clint/EdgeX/edgex-messagebus-client`

## 🚀 在新项目中使用

### 1. 添加依赖

在您的新项目中，添加模块依赖：

```bash
go mod init your-project-name
go get github.com/clint456/edgex-messagebus-client
```

### 2. 导入和使用

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
        "message": "Hello from external module!",
        "timestamp": time.Now(),
    }
    client.Publish("my/topic", data)

    // 订阅消息
    handler := func(topic string, message types.MessageEnvelope) error {
        fmt.Printf("收到消息: %s\n", string(message.Payload.([]byte)))
        return nil
    }
    client.SubscribeSingle("my/topic", handler)

    time.Sleep(5 * time.Second)
}
```

## 🔧 本地开发模式

如果您想在本地开发和测试模块，可以使用 `replace` 指令：

### 1. 在 go.mod 中添加本地路径

```go
module your-project

go 1.21

require (
    github.com/clint456/edgex-messagebus-client v0.1.0
    // 其他依赖...
)

replace github.com/clint456/edgex-messagebus-client => ../edgex-messagebus-client
```

### 2. 使用 -mod=mod 编译

```bash
go build -mod=mod ./...
```

## 📚 API 参考

### 主要类型

```go
// 配置结构
type Config struct {
    Host     string  // MQTT Broker 主机
    Port     int     // MQTT Broker 端口
    Protocol string  // 协议 (tcp, ssl, ws, wss)
    Type     string  // 消息总线类型 (mqtt, nats)
    ClientID string  // 客户端 ID
    Username string  // 用户名 (可选)
    Password string  // 密码 (可选)
    QoS      int     // QoS 级别
}

// 消息处理函数
type MessageHandler func(topic string, message types.MessageEnvelope) error
```

### 主要方法

| 方法 | 描述 | 示例 |
|------|------|------|
| `NewClient(config, logger)` | 创建客户端 | `client, err := messagebus.NewClient(config, lc)` |
| `Connect()` | 连接 | `client.Connect()` |
| `Disconnect()` | 断开连接 | `client.Disconnect()` |
| `Publish(topic, data)` | 发布消息 | `client.Publish("topic", data)` |
| `Subscribe(topics, handler)` | 订阅主题 | `client.Subscribe([]string{"topic"}, handler)` |
| `HealthCheck()` | 健康检查 | `client.HealthCheck()` |

## 🌟 优势

### 1. **模块化设计**
- 独立的 Go 模块，可在任何项目中使用
- 清晰的 API 接口
- 版本化管理

### 2. **易于维护**
- 单一职责原则
- 独立的测试和文档
- 便于更新和修复

### 3. **重用性强**
- 不需要复制代码
- 统一的 MessageBus 操作接口
- 支持多个项目同时使用

### 4. **开发友好**
- 完整的示例代码
- 详细的文档说明
- 本地开发支持

## 📝 发布到 GitHub

### 1. 创建 GitHub 仓库

```bash
cd /home/clint/EdgeX/edgex-messagebus-client
git init
git add .
git commit -m "Initial commit: EdgeX MessageBus Client module"
git remote add origin https://github.com/clint456/edgex-messagebus-client.git
git push -u origin main
```

### 2. 创建版本标签

```bash
git tag v0.1.0
git push origin v0.1.0
```

### 3. 在其他项目中使用

```bash
go get github.com/clint456/edgex-messagebus-client@v0.1.0
```

## 🔍 故障排除

### 常见问题

1. **模块找不到**
   - 确保模块已发布到 GitHub
   - 检查 go.mod 中的模块路径
   - 使用 `go mod tidy` 更新依赖

2. **本地开发问题**
   - 使用 `replace` 指令指向本地路径
   - 使用 `-mod=mod` 编译选项
   - 确保本地模块路径正确

3. **版本冲突**
   - 检查依赖版本兼容性
   - 使用 `go mod graph` 查看依赖关系
   - 必要时更新依赖版本

## 🎉 总结

通过将 MessageBus 客户端提取为独立模块，您现在可以：

- ✅ 在任何 Go 项目中轻松使用 EdgeX MessageBus 功能
- ✅ 避免代码重复和维护多个副本
- ✅ 享受模块化设计带来的便利
- ✅ 独立更新和维护 MessageBus 功能

这种模块化的方法是 Go 生态系统的最佳实践，让您的代码更加整洁和可维护！
