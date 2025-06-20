package interfaces

// DataPublisher 定义数据发布的通用接口，供 mqttbus、dataparse 等包依赖

type DataPublisher interface {
	PublishData(topic string, data []byte) error
}
