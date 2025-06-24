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
	"device-ble/pkg/config"
	"fmt"
	"log"

	"github.com/edgexfoundry/device-sdk-go/v4/pkg/startup"
)

const (
	serviceName string = "device-ble"
)

func main() {

	// 注入自定义配置
	cfg, err := config.Load("./res/configBle.yaml")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("串口配置:", cfg.Serial)
	fmt.Println("MQTT配置:", cfg.MQTT)

	d := driverpkg.Driver{
		Config: cfg,
	}

	startup.Bootstrap(serviceName, device.Version, &d)
}
