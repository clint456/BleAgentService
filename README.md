# BleAgentService
## branch
- v3.1  暂存已提交
> 支持napa版本
> 实现通过Edgex Ui控件触发蓝牙模块的基本操作、String、Json数据的收发
>

- v4.0  暂存提交
> 支持odassa版本
> - 实现v3.1的功能
> - 自动订阅edgex/events
> - 将消息转发至 自定义消息主题: edgex/HyData/events/+/+/+
> - 将消息下发至 Ble模块
> - 

（计划完成）监听来自手机与messageBus上用户（edgex/HyCommand/+/+/+）的命令请求，并发送回应至手机或messagebus上的用户（edgex/HyResponse/+/+/+）
