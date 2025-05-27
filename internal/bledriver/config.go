// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019-2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package bledriver

import (
	"fmt"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"
)

type ServiceConfig struct {
	MQTTBrokerInfo MQTTBrokerInfo
}

// UpdateFromRaw updates the service's full configuration from raw data received from
// the Service Provider.
func (sw *ServiceConfig) UpdateFromRaw(rawConfig interface{}) bool {
	configuration, ok := rawConfig.(*ServiceConfig)
	if !ok {
		return false //errors.New("unable to cast raw config to type 'ServiceConfig'")
	}

	*sw = *configuration

	return true
}

type MQTTBrokerInfo struct {
	Schema    string
	Host      string
	Port      int
	Qos       int
	KeepAlive int
	ClientId  string

	CredentialsRetryTime  int
	CredentialsRetryWait  int
	ConnEstablishingRetry int
	ConnRetryWaitTime     int

	AuthMode        string
	CredentialsName string

	IncomingTopic string
	ResponseTopic string

	Writable WritableInfo
}

// Validate ensures your custom configuration has proper values.
func (info *MQTTBrokerInfo) Validate() errors.EdgeX {
	if info.Writable.ResponseFetchInterval == 0 {
		return errors.NewCommonEdgeX(errors.KindContractInvalid, "MQTTBrokerInfo.Writable.ResponseFetchInterval configuration setting can not be blank", nil)
	}
	return nil
}

type WritableInfo struct {
	// ResponseFetchInterval specifies the retry interval(milliseconds) to fetch the command response from the MQTT broker
	ResponseFetchInterval int
}

func fetchCommandTopic(protocols map[string]models.ProtocolProperties) (string, errors.EdgeX) {
	properties, ok := protocols[Protocol]
	if !ok {
		return "", errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("'%s' protocol properties is not defined", Protocol), nil)
	}
	commandTopic, ok := properties[CommandTopic]
	if !ok {
		return "", errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("'%s' not found in the '%s' protocol properties", CommandTopic, Protocol), nil)
	}
	commandTopicString, ok := commandTopic.(string)
	if !ok {
		return "", errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("cannot convert '%v' to string type", CommandTopic), nil)
	}

	return commandTopicString, nil
}
