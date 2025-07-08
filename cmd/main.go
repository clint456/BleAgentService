// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2021 Jiangxing Intelligence Ltd
//
// SPDX-License-Identifier: Apache-2.0

// This package provides device service of a uart devices.
package main

import (
	device "device-ble"
	driverpkg "device-ble/internal/driver"

	"github.com/edgexfoundry/device-sdk-go/v4/pkg/startup"
)

const (
	serviceName string = "device-ble"
)

func main() {
	// 初始化Driver
	d := &driverpkg.Driver{}

	// 启动服务
	startup.Bootstrap(serviceName, device.Version, d)
}
