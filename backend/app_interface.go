package backend

import (
	"context"
	"time"
)

// AppInterface defines the interface that the main App struct must implement
type AppInterface interface {
	// Context management
	GetCtx() context.Context

	// Kafka methods that need to be implemented by the main App
	KafkaConnect(KafkaConfig) (string, error)
	KafkaDisconnect(string) error
	KafkaListTopics(string) ([]TopicInfo, error)
	KafkaStartConsumer(ConsumerConfig) (string, error)
	KafkaStopConsumer(string, string) error
	KafkaProduceMessage(ProducerConfig) error
	EmitStreamMessage(string, string, string, string)
}

// EmitStreamMessage is a regular function wrapper for the EmitStreamMessage method
func EmitStreamMessage(app AppInterface, connectionID, direction, protocol, payload string) {
	app.EmitStreamMessage(connectionID, direction, protocol, payload)
}

// StreamMessage contains a message in the stream
type StreamMessage struct {
	ID        string                 `json:"id"`
	Direction string                 `json:"direction"` // "inbound", "outbound", "system", "error"
	Protocol  string                 `json:"protocol"`
	Payload   string                 `json:"payload"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}