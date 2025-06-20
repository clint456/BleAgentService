package uart

import (
	"device-ble/internal/mocks"
	"testing"
	"time"
)

func TestSerialQueue_SendCommand_Mock(t *testing.T) {
	mock := &mocks.MockSerialQueue{
		SendCommandFunc: func(command []byte, timeout time.Duration) (string, error) {
			if string(command) != "test" {
				t.Errorf("unexpected command: %s", command)
			}
			return "OK", nil
		},
	}
	resp, err := mock.SendCommand([]byte("test"), time.Second)
	if err != nil || resp != "OK" {
		t.Fatalf("SendCommand failed: %v, resp: %s", err, resp)
	}
}

func TestSerialQueue_Close_Mock(t *testing.T) {
	called := false
	mock := &mocks.MockSerialQueue{
		CloseFunc: func() { called = true },
	}
	mock.Close()
	if !called {
		t.Error("CloseFunc was not called")
	}
}
