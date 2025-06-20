package ble

import (
	"device-ble/internal/mocks"
	"testing"
)

func TestBLEController_InitializeAsPeripheral_Mock(t *testing.T) {
	mock := &mocks.MockBLEController{
		InitializeAsPeripheralFunc: func() error { return nil },
	}
	if err := mock.InitializeAsPeripheral(); err != nil {
		t.Fatalf("InitializeAsPeripheral failed: %v", err)
	}
}

func TestBLEController_GetQueue_Mock(t *testing.T) {
	mock := &mocks.MockBLEController{
		GetQueueFunc: func() interface{} { return &mocks.MockSerialQueue{} },
	}
	if mock.GetQueue() == nil {
		t.Error("GetQueue should not return nil")
	}
}
