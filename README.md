# device-ble-go 项目说明

## 项目简介

`device-ble-go` 是基于 EdgeX Foundry v4 的 BLE 设备服务，专注于 BLE 设备与 EdgeX 微服务的高效集成。项目采用现代 Go 工程最佳实践，强调接口解耦、依赖注入、handler 闭包、单一职责和可测试性，适合生产环境和二次开发。

---

## 目录结构

```
project-root/
├── cmd/                # 程序入口（main.go/res 配置）
├── internal/
│   └── driver/         # 业务主流程（driver.go 等，部分接口也定义于此）
├── pkg/
│   ├── mqttbus/        # 消息总线实现（mqttClient.go）
│   └── dataparse/      # 数据处理实现（dataPublisher.go）
├── docs/               # 文档
├── vendor/             # 依赖
├── go.mod/go.sum       # Go模块定义
└── ...
```
> **接口建议**：目前接口类型主要定义在 internal/driver 或 pkg/ 下。推荐后续将所有跨包接口逐步迁移到 internal/interfaces 目录，以便解耦和团队协作。

---

## 架构与核心模块

### 1. 串口通信（pkg/uart）
- 负责与物理 BLE 模块的底层通信。
- 支持串口配置、命令队列、超时与错误处理。

### 2. BLE 控制器（pkg/ble）
- 封装 BLE 设备的命令、状态管理与数据收发。
- 提供标准化 AT 命令接口。

### 3. 消息总线（pkg/mqttbus）
- 基于自定义 messagebus 客户端，负责与 EdgeX 消息总线交互。
- 只实现接口，不依赖业务包，支持健康检查、自动重连。

### 4. 数据处理（pkg/dataparse）
- 负责 BLE 数据的格式转换、分包、发布到消息总线。
- 只依赖接口，便于扩展和测试。

### 5. 驱动主流程（internal/driver）
- 负责依赖注入、handler 组装、服务生命周期管理。
- 通过接口组合各核心模块，保持解耦。

---

## 现代 Go 工程最佳实践

- **接口优先**：所有跨包依赖均通过接口（如 MessageBusClient）实现，接口定义集中在 internal/interfaces。
- **依赖注入**：所有依赖通过构造函数、结构体字段或初始化方法注入，便于 mock 和单元测试。
- **handler 闭包**：handler 作为参数传递，闭包捕获依赖，便于扩展和测试。
- **单一职责**：每个包/模块只做一件事，便于维护和扩展。
- **只 import 接口包**：实现包绝不 import 其他实现包，彻底无循环依赖。
- **mock 测试**：所有接口都可 mock，便于隔离测试。

---

## 配置与运行

- 主配置文件：`cmd/res/configuration.yaml`，支持 MQTT、串口、BLE 等参数。
- 设备配置：`cmd/res/devices/devices.yaml`，定义设备属性与协议。
- 编译运行：
  ```bash
  go mod tidy
  make
  ./cmd/device-ble -o -d -cp
  ```
- 详细配置、API、调试、监控等见 docs/ 目录。

---

## 典型开发流程与扩展方式

1. **定义接口**（internal/interfaces/xxx.go）
2. **实现接口**（pkg/xxx/xxx.go，只 import 接口包）
3. **依赖注入**（driver.go/main.go 组装依赖，通过接口传递）
4. **handler 扩展**（通过闭包注入依赖，便于自定义业务逻辑）
5. **单元测试**（mock 接口，隔离测试）

### 代码示例

```go
// internal/interfaces/messagebus.go
 type MessageBusClient interface {
     Connect() error
     Publish(topic string, data interface{}) error
     // ...
 }

// pkg/mqttbus/mqttClient.go
 type EdgexMessageBusClient struct { ... }
 func NewEdgexMessageBusClient(..., handler func(topic string, envelope types.MessageEnvelope) error) (interfaces.MessageBusClient, error) { ... }

// internal/driver/driver.go
 handler := func(topic string, envelope types.MessageEnvelope) error {
     var data map[string]interface{}
     // ...
     dataparse.PublishToMessageBus(d.messageBus, data, topic)
     return nil
 }
```

---

## API 参考

### MessageBus 客户端
- `Connect()` 连接到消息总线
- `Disconnect()` 断开连接
- `Publish(topic, data)` 发布消息
- `Subscribe(topics, handler)` 订阅主题
- `HealthCheck()` 健康检查
- `GetClientInfo()` 获取客户端信息

### BLE 控制器
- `InitializeAsPeripheral()` 初始化外围设备
- `executeCommand()` 执行单个命令
- `sendCommandAndWaitResponse()` 发送命令并等待响应

---

## 常见问题与故障排查

- **编译失败**：`go mod tidy && make`，确保依赖完整。
- **串口连接失败**：检查设备路径、权限，参考 `docs/` 故障排查。
- **MQTT 连接失败**：检查网络、配置、Broker 状态。
- **BLE 初始化失败**：检查 AT 命令响应、硬件连接。
- **服务日志与监控**：`journalctl -u device-ble -f`，或查看 API 健康检查。

---

## 贡献指南

1. Fork 项目，创建功能分支。
2. 遵循接口优先、依赖注入、handler 闭包、单一职责等最佳实践。
3. 提交 Pull Request，附详细说明。
4. 代码需有注释、单元测试。

---

## 致谢

- [EdgeX Foundry](https://www.edgexfoundry.org/)
- [Jiangxing Intelligence](https://www.jiangxingai.com)
- HCL Technologies(EPL Team)
- 开源社区

---

## 路线图

- 完善单元测试与集成测试
- 性能与安全性增强
- 插件化与多设备支持
- 云端与 AI 集成

---

**如需详细开发模板、CI 配置、mock 示例等，请参考 docs/ 或联系维护者。**



