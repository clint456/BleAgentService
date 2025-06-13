# 使用指南

## 🚀 快速开始

### 安装

```bash
go get github.com/clint456/edgex-messagebus-client
```

### 基本使用

```go
package main

import (
    "log"
    "time"
    
    messagebus "github.com/clint456/edgex-messagebus-client"
    "github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
    "github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"
)

func main() {
    // 创建日志客户端
    lc := logger.NewClient("MyApp", "DEBUG")

    // 配置
    config := messagebus.Config{
        Host:     "localhost",
        Port:     1883,
        Protocol: "tcp",
        Type:     "mqtt",
        ClientID: "my-client",
    }

    // 创建客户端
    client, err := messagebus.NewClient(config, lc)
    if err != nil {
        log.Fatal(err)
    }

    // 连接
    if err := client.Connect(); err != nil {
        log.Fatal(err)
    }
    defer client.Disconnect()

    // 发布消息
    data := map[string]interface{}{
        "message": "Hello EdgeX!",
        "timestamp": time.Now(),
    }
    client.Publish("my/topic", data)

    // 订阅消息
    handler := func(topic string, message types.MessageEnvelope) error {
        log.Printf("收到消息: %s", string(message.Payload.([]byte)))
        return nil
    }
    client.SubscribeSingle("my/topic", handler)

    time.Sleep(5 * time.Second)
}
```

## 📚 详细示例

### 1. 发布不同类型的数据

```go
// JSON 数据
jsonData := map[string]interface{}{
    "temperature": 25.6,
    "humidity": 60.2,
}
client.Publish("sensors/data", jsonData)

// 字符串数据
client.Publish("sensors/status", "online")

// 二进制数据
binaryData := []byte{0x01, 0x02, 0x03}
client.PublishBinaryData("sensors/binary", binaryData)
```

### 2. 高级订阅

```go
// 订阅多个主题
topics := []string{"sensors/#", "devices/#", "events/#"}
handler := func(topic string, message types.MessageEnvelope) error {
    log.Printf("主题: %s, 消息: %s", topic, string(message.Payload.([]byte)))
    return nil
}
client.Subscribe(topics, handler)
```

### 3. 请求-响应模式

```go
// 创建请求
requestData := map[string]string{
    "command": "getStatus",
    "deviceId": "sensor01",
}
envelope, _ := client.CreateMessageEnvelope(requestData, "")

// 发送请求并等待响应
response, err := client.Request(
    envelope,
    "commands/request",
    "commands/response",
    5*time.Second,
)
if err != nil {
    log.Printf("请求失败: %v", err)
} else {
    log.Printf("响应: %s", string(response.Payload.([]byte)))
}
```

### 4. 错误处理

```go
// 监听错误
go func() {
    for err := range client.GetErrorChannel() {
        log.Printf("MessageBus 错误: %v", err)
    }
}()

// 健康检查
if err := client.HealthCheck(); err != nil {
    log.Printf("健康检查失败: %v", err)
}

// 获取客户端信息
info := client.GetClientInfo()
log.Printf("客户端状态: %+v", info)
```

## ⚙️ 配置选项

```go
config := messagebus.Config{
    Host:     "localhost",    // MQTT Broker 主机
    Port:     1883,          // MQTT Broker 端口
    Protocol: "tcp",         // 协议: tcp, ssl, ws, wss
    Type:     "mqtt",        // 类型: mqtt, nats
    ClientID: "my-client",   // 客户端 ID
    Username: "user",        // 用户名 (可选)
    Password: "pass",        // 密码 (可选)
    QoS:      1,            // QoS 级别: 0, 1, 2
}
```

## 🔧 最佳实践

1. **资源管理**: 始终调用 `Disconnect()`
2. **错误处理**: 检查所有操作的返回错误
3. **日志记录**: 使用适当的日志级别
4. **主题设计**: 使用有意义的主题层次结构
5. **消息格式**: 保持消息格式的一致性

## 🐛 故障排除

### 连接问题
- 检查 MQTT Broker 是否运行
- 验证主机地址和端口
- 检查网络连接

### 订阅无消息
- 确认主题名称正确
- 检查 QoS 设置
- 验证消息处理函数

### 发布失败
- 检查连接状态
- 验证消息格式
- 查看错误日志
