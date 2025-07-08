# v4.1.2更新内容
## 修改配置载入方式
`Uart`、`MqttClient`初始化载入配置文件改为`EdgeX标准读取方式`
## 新增ReStart资源
可重启该设备服务，刷新初始化配置

# 存在缺陷
## MessageBus请求性能瓶颈
对Mqtt broker请求间隔不能小于100ms，否者会出现Broker卡死的问题，后期建议更新为官方的[MessageBus库](https://github.com/edgexfoundry/go-mod-messaging.git)

