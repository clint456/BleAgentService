package interfaces

// BLEController 定义 BLE 控制器的通用接口
// 只暴露需要跨包调用的方法

type BLEController interface {
	InitializeAsPeripheral() error
	GetQueue() interface{} // 返回串口队列，具体类型由实现决定
}
