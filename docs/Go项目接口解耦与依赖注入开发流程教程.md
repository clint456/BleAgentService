# Go 项目接口解耦与依赖注入开发流程

---

## 1. 设计阶段

### 1.1 明确模块边界
- 每个包（package）只负责单一职责（如 mqttbus、dataparse、ble、driver）。
- 业务逻辑与基础设施（如消息总线、数据库、外部服务）分离。

### 1.2 先定义接口，再写实现
- 在 interface.go 或 internal/interfaces 包中定义所有跨包依赖的接口。
- 例如：
  ```go
  // internal/interfaces/messagebus.go
  package interfaces
  
  type MessageBusClient interface {
      Connect() error
      Disconnect() error
      IsConnected() bool
      Publish(topic string, data interface{}) error
      Subscribe(topics []string, handler func(topic string, envelope types.MessageEnvelope) error) error
  }
  ```

---

## 2. 编码阶段

### 2.1 只 import 必需的接口和标准库/三方库
- 绝不跨包 import 业务实现包，避免循环依赖。
- 只 import interface 包和标准库/三方库。

### 2.2 依赖注入
- 通过构造函数、初始化方法或结构体字段注入依赖。
- 例如：
  ```go
  type Driver struct {
      messageBus interfaces.MessageBusClient
      bleCtrl    interfaces.BLEController
  }
  ```

### 2.3 handler 闭包注入
- handler 作为参数传递，内部可闭包捕获依赖。
- 例如：
  ```go
  handler := func(topic string, envelope types.MessageEnvelope) error {
      // 可访问 driver.messageBus、driver.bleCtrl
      ...
      return nil
  }
  ```

### 2.4 只通过接口交互
- 业务逻辑只依赖接口，不关心具体实现。
- 便于 mock、单元测试和后续扩展。

---

## 3. 测试与维护

### 3.1 单元测试
- 使用 mock 实现接口，进行隔离测试。
- 例如：
  ```go
  type MockMessageBus struct { ... }
  func (m *MockMessageBus) Publish(...) error { ... }
  ```

### 3.2 代码审查
- 检查是否有跨包 import 实现包的情况。
- 检查 handler、依赖注入、接口解耦是否规范。

### 3.3 自动化工具
- 使用 goimports、golangci-lint 等工具自动清理未用 import、检测代码规范。
- 定期 go mod tidy 保持依赖整洁。

---

## 4. 典型开发流程示例

1. **定义接口**（internal/interfaces/xxx.go）
2. **实现接口**（如 driver/mqttbus/mqttClient.go 实现 MessageBusClient）
3. **在业务包中只依赖接口**（如 driver.go 只用 interfaces.MessageBusClient）
4. **通过依赖注入组装 handler 和依赖**（如 main.go 或 driver.go 中组装）
5. **handler 只通过接口调用业务方法**（如 PublishToMessageBus、SendToBlE）
6. **单元测试时 mock 接口，隔离测试业务逻辑**

---

## 5. 常见反例（应避免）

- 包 A import 包 B，包 B 又 import 包 A（循环依赖）。
- handler 直接依赖全局变量或具体实现。
- 业务逻辑直接 new 具体实现而不是通过接口。

---

## 6. 推荐工具

- [golangci-lint](https://golangci-lint.run/)：一站式 Go 代码静态检查工具
- [goimports](https://pkg.go.dev/golang.org/x/tools/cmd/goimports)：自动整理 import
- [mockery](https://github.com/vektra/mockery)：自动生成接口 mock
- [GoLand/VSCode Go 插件](https://github.com/golang/vscode-go)：IDE 支持

---

## 7. 总结

- **接口优先，解耦为王。**
- **依赖注入，灵活扩展。**
- **handler 闭包，便于测试。**
- **只 import 接口，绝不 import 实现。**
- **自动化工具保驾护航。**

---

# Go 项目接口解耦与依赖注入开发流程实战教程

---

## 1. 推荐目录结构

```
project-root/
├── cmd/                # 程序入口
│   └── main.go
├── internal/
│   ├── interfaces/     # 只放接口定义
│   │   └── messagebus.go
│   └── driver/         # 业务主流程
│       └── driver.go
├── pkg/                # 具体实现
│   ├── mqttbus/        # 消息总线实现
│   │   └── mqttClient.go
│   └── dataparse/      # 数据处理实现
│       └── dataPublisher.go
├── go.mod
└── ...
```

---

## 2. 典型接口定义（internal/interfaces/messagebus.go）

```go
package interfaces

import "github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"

type MessageBusClient interface {
    Connect() error
    Disconnect() error
    IsConnected() bool
    Publish(topic string, data interface{}) error
    Subscribe(topics []string, handler func(topic string, envelope types.MessageEnvelope) error) error
}
```

---

## 3. 具体实现（pkg/mqttbus/mqttClient.go）

```go
package mqttbus

import (
    "fmt"
    "strings"
    "github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"
    "github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
    messagebus "github.com/clint456/edgex-messagebus-client"
    "project-root/internal/interfaces"
)

type EdgexMessageBusClient struct {
    client *messagebus.Client
}

func NewEdgexMessageBusClient(cfg map[string]interface{}, logger logger.LoggingClient, subscribeTopics []string, handler func(topic string, envelope types.MessageEnvelope) error) (interfaces.MessageBusClient, error) {
    config := messagebus.Config{
        Host:     cfg["Host"].(string),
        Port:     cfg["Port"].(int),
        Protocol: strings.ToLower(cfg["Protocol"].(string)),
        Type:     "mqtt",
        ClientID: cfg["ClientID"].(string),
        QoS:      cfg["QoS"].(int),
        Username: cfg["Username"].(string),
        Password: cfg["Password"].(string),
    }
    client, err := messagebus.NewClient(config, logger)
    if err != nil {
        return nil, fmt.Errorf("创建MessageBus客户端失败: %w", err)
    }
    if err := client.Connect(); err != nil {
        return nil, fmt.Errorf("连接MessageBus失败: %w", err)
    }
    wrappedHandler := func(topic string, envelope types.MessageEnvelope) error {
        return handler(topic, envelope)
    }
    if err := client.Subscribe(subscribeTopics, wrappedHandler); err != nil {
        client.Disconnect()
        return nil, fmt.Errorf("订阅主题失败: %w", err)
    }
    return &EdgexMessageBusClient{client: client}, nil
}

func (e *EdgexMessageBusClient) Connect() error    { return e.client.Connect() }
func (e *EdgexMessageBusClient) Disconnect() error { return e.client.Disconnect() }
func (e *EdgexMessageBusClient) IsConnected() bool { return e.client.IsConnected() }
func (e *EdgexMessageBusClient) Publish(topic string, data interface{}) error {
    return e.client.Publish(topic, data)
}
func (e *EdgexMessageBusClient) Subscribe(topics []string, handler func(topic string, envelope types.MessageEnvelope) error) error {
    wrappedHandler := func(topic string, envelope types.MessageEnvelope) error {
        return handler(topic, envelope)
    }
    return e.client.Subscribe(topics, wrappedHandler)
}
```

---

## 4. handler 闭包与依赖注入（internal/driver/driver.go）

```go
package driver

import (
    "encoding/json"
    "project-root/internal/interfaces"
    "project-root/pkg/dataparse"
    "project-root/pkg/mqttbus"
    "github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"
)

type Driver struct {
    messageBus interfaces.MessageBusClient
    // 其他依赖...
}

func (d *Driver) Initialize() error {
    // 组装 handler，闭包捕获 d.messageBus
    handler := func(topic string, envelope types.MessageEnvelope) error {
        var data map[string]interface{}
        if err := json.Unmarshal(envelope.Payload.([]byte), &data); err != nil {
            // 日志...
            return err
        }
        if err := dataparse.PublishToMessageBus(d.messageBus, data, topic); err != nil {
            // 日志...
            return err
        }
        // 其他业务...
        return nil
    }

    // 依赖注入
    cfg := map[string]interface{}{ /* ... */ }
    subscribeTopics := []string{"edgex/events/#"}
    msgBus, err := mqttbus.NewEdgexMessageBusClient(cfg, /* logger */, subscribeTopics, handler)
    if err != nil {
        return err
    }
    d.messageBus = msgBus
    return nil
}
```

---

## 5. 单元测试与 Mock

```go
type MockMessageBus struct{}
func (m *MockMessageBus) Connect() error { return nil }
func (m *MockMessageBus) Publish(topic string, data interface{}) error { return nil }
// ...实现接口所有方法

func TestDriverHandler(t *testing.T) {
    driver := &Driver{messageBus: &MockMessageBus{}}
    // 测试 handler 逻辑
}
```

---

## 6. 团队协作建议

- **接口优先**：所有跨包依赖先定义接口，后写实现。
- **只 import 接口包**：实现包绝不 import 其他实现包。
- **handler 只通过依赖注入传递**，不直接访问全局变量。
- **代码审查重点**：是否有跨包 import 实现包、handler 是否闭包注入依赖。
- **自动化工具**：CI 中强制 goimports、golangci-lint、go mod tidy。

---

## 7. 常见问题与应对

- **循环依赖**：检查 import 路径，接口单独放 internal/interfaces，所有实现包只 import 接口包。
- **handler 依赖全局变量**：重构为通过结构体字段或参数传递依赖。
- **mock 难写**：接口粒度过大，拆分为更细的接口。

---

## 8. 总结

- 目录结构清晰，接口与实现分离。
- 依赖注入和 handler 闭包让业务灵活、可测。
- 只 import 接口，彻底无循环依赖。
- 团队协作有规范，自动化工具保驾护航。

---

如需更详细的代码模板、CI 配置或团队文档范例，请随时告知！