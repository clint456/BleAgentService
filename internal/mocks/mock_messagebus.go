package mocks

// MockMessageBusClient 实现 interfaces.MessageBusClient
// 可通过函数变量自定义行为

type MockMessageBusClient struct {
	PublishFunc   func(topic string, payload []byte) error
	SubscribeFunc func(topic string, handler func([]byte)) error
}

func (m *MockMessageBusClient) Publish(topic string, payload []byte) error {
	if m.PublishFunc != nil {
		return m.PublishFunc(topic, payload)
	}
	return nil
}

func (m *MockMessageBusClient) Subscribe(topic string, handler func([]byte)) error {
	if m.SubscribeFunc != nil {
		return m.SubscribeFunc(topic, handler)
	}
	return nil
}
