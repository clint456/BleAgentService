package mocks

// MockBLEController 实现 interfaces.BLEController

type MockBLEController struct {
	InitializeAsPeripheralFunc func() error
	GetQueueFunc               func() interface{}
}

func (m *MockBLEController) InitializeAsPeripheral() error {
	if m.InitializeAsPeripheralFunc != nil {
		return m.InitializeAsPeripheralFunc()
	}
	return nil
}

func (m *MockBLEController) GetQueue() interface{} {
	if m.GetQueueFunc != nil {
		return m.GetQueueFunc()
	}
	return nil
}
