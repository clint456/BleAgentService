package driver

import (
	"github.com/edgexfoundry/device-sdk-go/v3/pkg/interfaces"
	dsModels "github.com/edgexfoundry/device-sdk-go/v3/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
)

type BleDriver struct {
	sdk      interfaces.DeviceServiceSDK
	lc       logger.LoggingClient
	asyncCh  chan<- *dsModels.AsyncValues
	deviceCh chan<- []dsModels.DiscoveredDevice
	uart     map[string]*Uart
}

func (s *BleDriver) Initialize(sdk interfaces.DeviceServiceSDK) error {
	s.sdk = sdk
	s.lc = sdk.LoggingClient()
	s.asyncCh = sdk.AsyncValuesChannel()
	s.deviceCh = sdk.DiscoveredDeviceChannel()

	s.uart = make(map[string]*Uart)

	return nil
}
