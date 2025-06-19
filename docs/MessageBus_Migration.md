# MessageBus 库迁移指南

## 概述

本文档描述了将 BLE Agent Service 中的 MQTT 监听客户端从 `paho.mqtt.golang` 迁移到自定义的 `github.com/clint456/edgex-messagebus-client` 库的过程和变化。

## 迁移原因

1. **统一接口**：使用统一的 MessageBus 接口，简化代码维护
2. **更好的集成**：与 EdgeX 生态系统更好的集成
3. **增强功能**：提供更丰富的功能，如健康检查、客户端信息等
4. **类型安全**：更好的类型安全和错误处理

## 主要变化

### 1. 依赖变化

**之前：**
```go
import (
    mqtt "github.com/eclipse/paho.mqtt.golang"
)
```

**现在：**
```go
import (
    messagebus "github.com/clint456/edgex-messagebus-client"
)
```

### 2. 客户端类型变化

**之前：**
```go
type Driver struct {
    mqttClient mqtt.Client  // 监听客户端
    // ...
}
```

**现在：**
```go
type Driver struct {
    mqttClient *messagebus.Client  // 监听客户端 - 使用新的messagebus库
    // ...
}
```

### 3. 客户端创建方式

**之前：**
```go
func (s *Driver) createMqttClient(serviceConfig *ServiceConfig) (mqtt.Client, errors.EdgeX) {
    opts := mqtt.NewClientOptions()
    opts.AddBroker(fmt.Sprintf("%s://%s", uri.Scheme, uri.Host))
    opts.SetClientID(clientID)
    // ... 更多配置
    
    client := mqtt.NewClient(opts)
    token := client.Connect()
    // ...
}
```

**现在：**
```go
func (s *Driver) createMessageBusClient(serviceConfig *ServiceConfig) (*messagebus.Client, errors.EdgeX) {
    config := messagebus.Config{
        Host:     serviceConfig.MQTTBrokerInfo.Host,
        Port:     serviceConfig.MQTTBrokerInfo.Port,
        Protocol: strings.ToLower(serviceConfig.MQTTBrokerInfo.Schema),
        Type:     "mqtt",
        ClientID: serviceConfig.MQTTBrokerInfo.ClientId,
        QoS:      serviceConfig.MQTTBrokerInfo.Qos,
    }
    
    client, err := messagebus.NewClient(config, s.lc)
    if err := client.Connect(); err != nil {
        return nil, err
    }
    // ...
}
```

### 4. 消息处理函数

**之前：**
```go
func (s *Driver) onIncomingDataReceived(client mqtt.Client, message mqtt.Message) {
    incomingTopic := message.Topic()
    payload := message.Payload()
    // ...
}
```

**现在：**
```go
func (s *Driver) onMessageBusDataReceived(topic string, message types.MessageEnvelope) error {
    // 处理不同类型的payload
    var payloadBytes []byte
    switch payload := message.Payload.(type) {
    case []byte:
        payloadBytes = payload
    case string:
        payloadBytes = []byte(payload)
    default:
        if data, ok := payload.(map[string]interface{}); ok {
            asyncData = data
        }
    }
    // ...
    return nil
}
```

### 5. 订阅方式

**之前：**
```go
token := client.Subscribe(incomingTopic, qos, s.onIncomingDataReceived)
```

**现在：**
```go
subscribeTopics := []string{incomingTopic}
err := client.Subscribe(subscribeTopics, s.onMessageBusDataReceived)
```

### 6. 客户端断开连接

**之前：**
```go
client.Disconnect(0)
```

**现在：**
```go
client.Disconnect()
```

## 新增功能

### 1. 健康检查
```go
if err := client.HealthCheck(); err != nil {
    fmt.Printf("健康检查失败: %v\n", err)
}
```

### 2. 客户端信息
```go
info := client.GetClientInfo()
fmt.Printf("客户端信息: %+v\n", info)
```

### 3. 更好的错误处理
新库提供了更详细的错误信息和更好的错误处理机制。

## 配置兼容性

现有的配置文件 `cmd/res/configuration.yaml` 中的 `MQTTBrokerInfo` 配置保持不变，新库会自动适配这些配置。

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
```

## 测试和验证

### 1. 运行示例
```bash
cd examples
go run messagebus_example.go
```

### 2. 编译项目
```bash
make
```

### 3. 运行服务
```bash
./cmd/device-ble -o -d -cp
```

## 注意事项

1. **向后兼容性**：转发客户端仍然使用 `go-mod-messaging`，保持与 EdgeX 的兼容性
2. **错误处理**：新的消息处理函数需要返回 error，确保正确处理错误
3. **类型处理**：新库对 payload 类型处理更加灵活，支持多种数据类型
4. **连接管理**：新库提供了更好的连接管理和自动重连功能

## 迁移检查清单

- [x] 更新导入语句
- [x] 修改客户端类型定义
- [x] 重构客户端创建函数
- [x] 更新消息处理函数
- [x] 修改订阅方式
- [x] 更新断开连接方法
- [x] 测试编译
- [x] 创建示例代码
- [x] 更新文档

## 故障排除

### 常见问题

1. **编译错误**：确保已正确安装新的 messagebus 库
   ```bash
   go get github.com/clint456/edgex-messagebus-client
   go mod tidy
   ```

2. **连接失败**：检查 MQTT Broker 配置和网络连接

3. **消息接收问题**：确认主题订阅正确，检查消息格式

### 调试技巧

1. 启用 DEBUG 日志级别
2. 使用健康检查功能验证连接状态
3. 检查客户端信息确认配置正确

## 总结

通过迁移到新的 messagebus 库，BLE Agent Service 获得了更好的功能性、可维护性和与 EdgeX 生态系统的集成。新库提供了更丰富的功能和更好的错误处理，同时保持了配置的向后兼容性。
