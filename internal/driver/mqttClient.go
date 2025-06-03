package driver

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-messaging/v4/messaging"
	"github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"
	"github.com/google/uuid"
)

func (s *Driver) initialMqttClient() error {
	// 初始化监听客户端
	s.serviceConfig = &ServiceConfig{}
	if err := s.sdk.LoadCustomConfig(s.serviceConfig, CustomConfigSectionName); err != nil {
		return fmt.Errorf("❌ 加载MQTTClint '%s' 自定义配置失败: %s", CustomConfigSectionName, err.Error())
	}
	s.lc.Debugf("✅️ MQTTClient自定义配置加载成功: %v", s.serviceConfig)
	if err := s.serviceConfig.MQTTBrokerInfo.Validate(); err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	if err := s.sdk.ListenForCustomConfigChanges(
		&s.serviceConfig.MQTTBrokerInfo.Writable,
		WritableInfoSectionName, s.updateWritableConfig); err != nil {
		return errors.NewCommonEdgeX(errors.Kind(err), fmt.Sprintf("‼❌️ 监听MQTTClint失败 '%s' 自定义配置改动", WritableInfoSectionName), err)
	}

	client, err := s.createMqttClient(s.serviceConfig)
	if err != nil {
		return errors.NewCommonEdgeX(errors.Kind(err), "‼❌️ 初始化MqttClient失败", err)
	}
	s.mqttClient = client

	// 初始化转发客户端
	s.transmitClient, err = s.NewMessageBusClient("tainsmitCient")
	return nil
}

func (s *Driver) updateWritableConfig(rawWritableConfig interface{}) {
	updated, ok := rawWritableConfig.(*WritableInfo)
	if !ok {
		s.lc.Error("❌ 更新writeable配置失败：不能将config源数据反射为'WritableInfo'")
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
			s.lc.Warnf("🔴 连接Mqtt代理服务器失败, %s, retrying", err)
			time.Sleep(time.Duration(serviceConfig.MQTTBrokerInfo.ConnEstablishingRetry) * time.Second)
			continue
		}
		break
	}
	return client, nil
}

func (s *Driver) getMqttClient(clientID string, uri *url.URL, keepAlive int) (mqtt.Client, error) {
	s.lc.Infof("⏩️ 创建Mqtt客户端并连接中: hostname=%v clientID=%v ", uri.Hostname(), clientID)
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
	incomingTopic := s.serviceConfig.MQTTBrokerInfo.IncomingTopic

	token := client.Subscribe(incomingTopic, qos, s.onIncomingDataReceived)
	if token.Wait() && token.Error() != nil {
		client.Disconnect(0)
		s.lc.Errorf("❌️ 不能订阅到'%s'主题: %s",
			incomingTopic, token.Error().Error())
		return
	}
	s.lc.Infof("📶 成功订阅到 '%s' 用于接收同步", incomingTopic)

}

func (s *Driver) NewMessageBusClient(ClientID string) (messaging.MessageClient, errors.EdgeX) {
	messageBus, err := messaging.NewMessageClient(types.MessageBusConfig{
		Broker: types.HostInfo{
			Host:     s.serviceConfig.MQTTBrokerInfo.Host,
			Port:     s.serviceConfig.MQTTBrokerInfo.Port,
			Protocol: s.serviceConfig.MQTTBrokerInfo.Schema,
		},
		Type: "mqtt",
		Optional: map[string]string{
			"ClientID": ClientID + uuid.New().String()},
	})

	if err != nil {
		return nil, errors.NewCommonEdgeXWrapper(fmt.Errorf("⛔️ 消息客户端失败: %v", err))
	}
	if messageBus == nil {
		return nil, errors.NewCommonEdgeXWrapper(fmt.Errorf("⛔️ 消息客户端为 nil"))
	}
	// 连接到 Broker
	if err := messageBus.Connect(); err != nil {
		return nil, errors.NewCommonEdgeXWrapper(fmt.Errorf("⛔️ 连接到 MQTT Broker 失败: %v", err))
	}
	s.lc.Debugf("✅️ %v 消息客户端初始化成功", ClientID)
	return messageBus, nil
}
