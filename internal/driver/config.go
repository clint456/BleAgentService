// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019-2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"fmt"
	"net/url"

	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/secret"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/config"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"
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

func (d *Driver) SetCredentials(uri *url.URL, secretProvider interfaces.SecretProvider, category string, authMode string, secretName string) error {
	switch authMode {
	case AuthModeUsernamePassword:
		credentials, err := d.GetCredentials(secretProvider, secretName)
		if err != nil {
			return fmt.Errorf("获取%s MQTT证书失败 '%s': %w", category, secretName, err)
		}

		d.logger.Infof("%s MQTT认证已加载", category)
		uri.User = url.UserPassword(credentials.Username, credentials.Password)

	case AuthModeNone:
		return nil
	default:
		return fmt.Errorf("MQTT %s 连接认证模式 '%s' 不支持", category, authMode)
	}

	return nil
}

func (d *Driver) GetCredentials(secretProvider interfaces.SecretProvider, secretName string) (config.Credentials, error) {
	credentials := config.Credentials{}

	timer := startup.NewTimer(d.serviceConfig.MQTTBrokerInfo.CredentialsRetryTime, d.serviceConfig.MQTTBrokerInfo.CredentialsRetryWait)

	var secretData map[string]string
	var err error
	for timer.HasNotElapsed() {
		secretData, err = secretProvider.GetSecret(secretName, secret.UsernameKey, secret.PasswordKey)
		if err == nil {
			break
		}

		d.logger.Warnf(
			"无法从SecretProvider获取MQTT证书 '%s': %s. 重试中，剩余时间: %s",
			secretName,
			err.Error(),
			timer.RemainingAsString())
		timer.SleepForInterval()
	}

	if err != nil {
		return credentials, err
	}

	credentials.Username = secretData[secret.UsernameKey]
	credentials.Password = secretData[secret.PasswordKey]

	return credentials, nil
}
