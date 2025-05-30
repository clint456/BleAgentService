package driver

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
)

func (s *Driver) initalMqttClient() error {
	s.serviceConfig = &ServiceConfig{}
	if err := s.sdk.LoadCustomConfig(s.serviceConfig, CustomConfigSectionName); err != nil {
		return fmt.Errorf("❌加载MQTTClint '%s' 自定义配置失败: %s", CustomConfigSectionName, err.Error())
	}
	s.lc.Debugf("✌️MQTTClient自定义配置加载成功: %v", s.serviceConfig)
	if err := s.serviceConfig.MQTTBrokerInfo.Validate(); err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	if err := s.sdk.ListenForCustomConfigChanges(
		&s.serviceConfig.MQTTBrokerInfo.Writable,
		WritableInfoSectionName, s.updateWritableConfig); err != nil {
		return errors.NewCommonEdgeX(errors.Kind(err), fmt.Sprintf("‼️不能监听MQTTClint '%s' 自定义配置改动", WritableInfoSectionName), err)
	}

	client, err := s.createMqttClient(s.serviceConfig)
	if err != nil {
		return errors.NewCommonEdgeX(errors.Kind(err), "‼️不能初始化MqttClient", err)
	}
	s.mqttClient = client
	return nil
}

func (s *Driver) updateWritableConfig(rawWritableConfig interface{}) {
	updated, ok := rawWritableConfig.(*WritableInfo)
	if !ok {
		s.lc.Error("❌unable to update writable config: Can not cast raw config to type 'WritableInfo'")
		return
	}
	s.serviceConfig.MQTTBrokerInfo.Writable = *updated
}

func (s *Driver) createMqttClient(serviceConfig *ServiceConfig) (mqtt.Client, errors.EdgeX) {
	var scheme = serviceConfig.MQTTBrokerInfo.Schema
	var brokerUrl = serviceConfig.MQTTBrokerInfo.Host
	var brokerPort = serviceConfig.MQTTBrokerInfo.Port
	var authMode = serviceConfig.MQTTBrokerInfo.AuthMode
	var secretName = serviceConfig.MQTTBrokerInfo.CredentialsName
	var mqttClientId = serviceConfig.MQTTBrokerInfo.ClientId
	var keepAlive = serviceConfig.MQTTBrokerInfo.KeepAlive

	uri := &url.URL{
		Scheme: strings.ToLower(scheme),
		Host:   fmt.Sprintf("%s:%d", brokerUrl, brokerPort),
	}

	err := s.SetCredentials(uri, s.sdk.SecretProvider(), "init", authMode, secretName)
	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(err)
	}

	var client mqtt.Client
	for i := 0; i <= serviceConfig.MQTTBrokerInfo.ConnEstablishingRetry; i++ {
		client, err = s.getMqttClient(mqttClientId, uri, keepAlive)
		if err != nil && i >= serviceConfig.MQTTBrokerInfo.ConnEstablishingRetry {
			return nil, errors.NewCommonEdgeXWrapper(err)
		} else if err != nil {
			s.lc.Warnf("‼️Unable to connect to MQTT broker, %s, retrying", err)
			time.Sleep(time.Duration(serviceConfig.MQTTBrokerInfo.ConnEstablishingRetry) * time.Second)
			continue
		}
		break
	}
	return client, nil
}

func (s *Driver) getMqttClient(clientID string, uri *url.URL, keepAlive int) (mqtt.Client, error) {
	s.lc.Infof("⏩️Create MQTT client and connection: hostname=%v clientID=%v ", uri.Hostname(), clientID)
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("%s://%s", uri.Scheme, uri.Host))
	opts.SetClientID(clientID)
	opts.SetUsername(uri.User.Username())
	password, _ := uri.User.Password()
	opts.SetPassword(password)
	opts.SetKeepAlive(time.Second * time.Duration(keepAlive))
	opts.SetAutoReconnect(true)
	opts.OnConnect = s.onConnectHandler

	client := mqtt.NewClient(opts)
	token := client.Connect()
	if token.Wait() && token.Error() != nil {
		return client, token.Error()
	}

	return client, nil
}

func (s *Driver) onConnectHandler(client mqtt.Client) {
	qos := byte(s.serviceConfig.MQTTBrokerInfo.Qos)
	responseTopic := s.serviceConfig.MQTTBrokerInfo.ResponseTopic
	incomingTopic := s.serviceConfig.MQTTBrokerInfo.IncomingTopic

	token := client.Subscribe(incomingTopic, qos, s.onIncomingDataReceived)
	if token.Wait() && token.Error() != nil {
		client.Disconnect(0)
		s.lc.Errorf("‼️could not subscribe to topic '%s': %s",
			incomingTopic, token.Error().Error())
		return
	}
	s.lc.Infof("📶订阅到'%s' 用于接收同步", incomingTopic)

	token = client.Subscribe(responseTopic, qos, s.onCommandResponseReceived)
	if token.Wait() && token.Error() != nil {
		client.Disconnect(0)
		s.lc.Errorf("could not subscribe to topic '%s': %s",
			responseTopic, token.Error().Error())
		return
	}
	s.lc.Infof("📶Subscribed to topic '%s' for receiving the request response", responseTopic)

}
