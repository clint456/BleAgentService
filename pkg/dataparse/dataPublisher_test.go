package dataparse

import (
	"device-ble/internal/mocks"
	"testing"
	"time"
)

func TestPublishToMessageBus(t *testing.T) {
	called := false
	mockBus := &mocks.MockMessageBusClient{
		PublishFunc: func(topic string, payload []byte) error {
			called = true
			if topic != "test/topic" {
				t.Errorf("unexpected topic: %s", topic)
			}
			return nil
		},
	}
	err := PublishToMessageBus(mockBus, map[string]interface{}{"foo": "bar"}, "test/topic")
	if err != nil {
		t.Fatalf("PublishToMessageBus failed: %v", err)
	}
	if !called {
		t.Error("PublishFunc was not called")
	}
}

func TestSendToBlE(t *testing.T) {
	called := false
	mockQueue := &mocks.MockSerialQueue{
		SendCommandFunc: func(command []byte, timeout time.Duration) (string, error) {
			called = true
			return "OK\n", nil
		},
	}
	mockBle := &mocks.MockBLEController{
		GetQueueFunc: func() interface{} { return mockQueue },
	}
	SendToBlE(mockBle, map[string]interface{}{"foo": "bar"})
	if !called {
		t.Error("SendCommandFunc was not called")
	}
}
