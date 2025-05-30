// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2021 Jiangxing Intelligence Ltd
//
// SPDX-License-Identifier: Apache-2.0

// This package provides device service of a uart devices.
package main

import (
	"internal/driver"

	"github.com/edgexfoundry/device-sdk-go/v4/pkg/startup"
)

const (
	serviceName string = "ble-agent-service"
)

func main() {
	d := driver.Driver{}
	startup.Bootstrap(serviceName, driver.Version, &d)
}
