# v4.1.1更新内容
## 1. 依赖注入与关注点分离
Driver 只负责协调和生命周期管理，不再直接实现业务逻辑。
CommandService 和 AgentService 独立为单独文件（service.go），专注于命令分发和透明代理数据处理，便于扩展和测试。
所有依赖（logger、配置、串口、队列、BLE、消息总线、Service等）都通过 main.go 装配层初始化并注入，Driver/Service 只持有接口，不负责 new 依赖。
## 2. 回调闭包与上下文保留
串口队列的回调依然通过闭包方式注册，Driver 的上下文和状态不会丢失，功能完全兼容。
## 3. 业务逻辑迁移与保留
原有的命令处理、透明代理数据处理等业务逻辑全部迁移到 Service 层，功能未丢失。
Driver 的回调只做转发，Service 负责具体业务实现。
## 4. 装配层（main.go）标准化
main.go 负责所有依赖的初始化和装配，流程清晰、易于维护和测试。
BLE 设备初始化（InitializeAsPeripheral）已在 main.go 明确调用，保证设备启动流程完整。