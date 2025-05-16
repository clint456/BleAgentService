// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2021 Jiangxing Intelligence Ltd
//
// SPDX-License-Identifier: Apache-2.0

// This package provides device service of a uart devices.
package main

import (
	"github.com/edgexfoundry/device-sdk-go/v3/pkg/startup"

	"github.com/edgexfoundry/device-uart"
	"github.com/edgexfoundry/device-uart/internal/driver"
)

const (
	serviceName string = "device-ble"
)

func main() {
	d := driver.BleDriver{}
	startup.Bootstrap(serviceName, device.Version, &d)
}
