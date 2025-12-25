package backend

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/segmentio/kafka-go/sasl/scram"
)

type KafkaConnection struct {
	ID        string
	Brokers   []string
	Config    *KafkaConfig
	Dialer    *kafka.Dialer
	Consumers map[string]*ConsumerInstance
	mu        sync.RWMutex
}

type ConsumerInstance struct {
	ID            string
	Topic         string
	ConsumerGroup string
	Reader        *kafka.Reader
	Cancel        context.CancelFunc
	IsActive      bool
}

type KafkaConfig struct {
	BootstrapServers  []string `json:"bootstrapServers"`
	ClientID          string   `json:"clientId"`
	AuthMechanism     string   `json:"authMechanism"`
	SaslUsername      string   `json:"saslUsername"`
	SaslPassword      string   `json:"saslPassword"`
	UseTLS            bool     `json:"useTLS"`
	TLSSkipVerify     bool     `json:"tlsSkipVerify"`
	ConnectionTimeout int      `json:"connectionTimeout"`
}

type TopicInfo struct {
	Name       string `json:"name"`
	Partitions int    `json:"partitions"`
}

type ConsumerConfig struct {
	ConnectionID   string `json:"connectionId"`
	Topic          string `json:"topic"`
	Partitions     []int  `json:"partitions"`
	ConsumerGroup  string `json:"consumerGroup"`
	OffsetStrategy string `json:"offsetStrategy"`
	CustomOffset   int64  `json:"customOffset"`
	AutoCommit     bool   `json:"autoCommit"`
}

type ProducerConfig struct {
	ConnectionID string            `json:"connectionId"`
	Topic        string            `json:"topic"`
	Partition    int               `json:"partition"`
	Key          string            `json:"key"`
	Value        string            `json:"value"`
	Headers      map[string]string `json:"headers"`
	Compression  string            `json:"compression"`
	Acks         int               `json:"acks"`
}

var (
	kafkaConnections = make(map[string]*KafkaConnection)
	kafkaMutex       sync.RWMutex
)

func KafkaConnect(app AppInterface, config KafkaConfig) (string, error) {
	log.Printf("[Kafka] Connecting to brokers: %v", config.BootstrapServers)

	connectionID := uuid.New().String()

	dialer := &kafka.Dialer{
		Timeout:   time.Duration(config.ConnectionTimeout) * time.Millisecond,
		DualStack: true,
		ClientID:  config.ClientID,
	}

	if config.UseTLS {
		dialer.TLS = &tls.Config{
			InsecureSkipVerify: config.TLSSkipVerify,
		}
	}

	if config.AuthMechanism != "none" && config.AuthMechanism != "" {
		mechanism, err := createSASLMechanism(config.AuthMechanism, config.SaslUsername, config.SaslPassword)
		if err != nil {
			return "", fmt.Errorf("failed to create SASL mechanism: %w", err)
		}
		dialer.SASLMechanism = mechanism
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := dialer.DialContext(ctx, "tcp", config.BootstrapServers[0])
	if err != nil {
		return "", fmt.Errorf("failed to connect to broker: %w", err)
	}
	defer conn.Close()

	_, err = conn.Brokers()
	if err != nil {
		return "", fmt.Errorf("failed to get cluster metadata: %w", err)
	}

	kafkaMutex.Lock()
	kafkaConnections[connectionID] = &KafkaConnection{
		ID:        connectionID,
		Brokers:   config.BootstrapServers,
		Config:    &config,
		Dialer:    dialer,
		Consumers: make(map[string]*ConsumerInstance),
	}
	kafkaMutex.Unlock()

	log.Printf("[Kafka] Connected successfully with ID: %s", connectionID)

	emitStreamMessage(app, connectionID, "system", "kafka", fmt.Sprintf("Connected to Kafka cluster: %v", config.BootstrapServers))

	return connectionID, nil
}

func KafkaDisconnect(app AppInterface, connectionID string) error {
	log.Printf("[Kafka] Disconnecting: %s", connectionID)

	kafkaMutex.Lock()
	conn, exists := kafkaConnections[connectionID]
	if !exists {
		kafkaMutex.Unlock()
		return fmt.Errorf("connection not found: %s", connectionID)
	}
	delete(kafkaConnections, connectionID)
	kafkaMutex.Unlock()

	conn.mu.Lock()
	for _, consumer := range conn.Consumers {
		if consumer.IsActive && consumer.Cancel != nil {
			consumer.Cancel()
		}
		if consumer.Reader != nil {
			consumer.Reader.Close()
		}
	}
	conn.mu.Unlock()

	log.Printf("[Kafka] Disconnected: %s", connectionID)
	return nil
}

func KafkaListTopics(app AppInterface, connectionID string) ([]TopicInfo, error) {
	kafkaMutex.RLock()
	conn, exists := kafkaConnections[connectionID]
	kafkaMutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("connection not found: %s", connectionID)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	kafkaConn, err := conn.Dialer.DialContext(ctx, "tcp", conn.Brokers[0])
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	defer kafkaConn.Close()

	partitions, err := kafkaConn.ReadPartitions()
	if err != nil {
		return nil, fmt.Errorf("failed to read partitions: %w", err)
	}

	topicMap := make(map[string]int)
	for _, p := range partitions {
		topicMap[p.Topic]++
	}

	topics := make([]TopicInfo, 0, len(topicMap))
	for name, partitionCount := range topicMap {
		topics = append(topics, TopicInfo{
			Name:       name,
			Partitions: partitionCount,
		})
	}

	log.Printf("[Kafka] Found %d topics", len(topics))
	return topics, nil
}

func KafkaStartConsumer(app AppInterface, config ConsumerConfig) (string, error) {
	log.Printf("[Kafka] Starting consumer for topic: %s", config.Topic)
	log.Printf("[Kafka] Received ConsumerGroup value: '%s' (length: %d)", config.ConsumerGroup, len(config.ConsumerGroup))

	kafkaMutex.RLock()
	conn, exists := kafkaConnections[config.ConnectionID]
	kafkaMutex.RUnlock()

	if !exists {
		return "", fmt.Errorf("connection not found: %s", config.ConnectionID)
	}

	consumerID := uuid.New().String()

	var startOffset int64
	switch config.OffsetStrategy {
	case "earliest":
		startOffset = kafka.FirstOffset
	case "latest":
		startOffset = kafka.LastOffset
	case "custom":
		startOffset = config.CustomOffset
	default:
		startOffset = kafka.LastOffset
	}

	readerConfig := kafka.ReaderConfig{
		Brokers:     conn.Brokers,
		Topic:       config.Topic,
		Partition:   0, // Always read from partition 0
		StartOffset: startOffset,
		Dialer:      conn.Dialer,
		MaxWait:     500 * time.Millisecond,
		MinBytes:    1,
		MaxBytes:    10e6, // 10MB
	}

	if len(config.Partitions) > 0 {
		readerConfig.Partition = config.Partitions[0]
	}

	log.Printf("[Kafka] Consumer config - Topic: %s, Partition: %d, StartOffset: %d (NO CONSUMER GROUP)",
		config.Topic, readerConfig.Partition, startOffset)

	reader := kafka.NewReader(readerConfig)

	ctx, cancel := context.WithCancel(context.Background())

	conn.mu.Lock()
	conn.Consumers[consumerID] = &ConsumerInstance{
		ID:            consumerID,
		Topic:         config.Topic,
		ConsumerGroup: config.ConsumerGroup,
		Reader:        reader,
		Cancel:        cancel,
		IsActive:      true,
	}
	conn.mu.Unlock()

	go consumeMessages(app, ctx, config.ConnectionID, consumerID, reader)

	log.Printf("[Kafka] Consumer started: %s for topic: %s (partition mode, no groups)", consumerID, config.Topic)
	emitStreamMessage(app, config.ConnectionID, "system", "kafka", fmt.Sprintf("Started consumer %s for topic: %s (partition: %d)", consumerID[:8], config.Topic, readerConfig.Partition))

	return consumerID, nil
}

func KafkaStopConsumer(app AppInterface, connectionID string, consumerID string) error {
	log.Printf("[Kafka] Stopping consumer: %s", consumerID)

	kafkaMutex.RLock()
	conn, exists := kafkaConnections[connectionID]
	kafkaMutex.RUnlock()

	if !exists {
		return fmt.Errorf("connection not found: %s", connectionID)
	}

	conn.mu.Lock()
	consumer, exists := conn.Consumers[consumerID]
	if !exists {
		conn.mu.Unlock()
		return fmt.Errorf("consumer not found: %s", consumerID)
	}

	if !consumer.IsActive {
		conn.mu.Unlock()
		return fmt.Errorf("consumer already stopped")
	}

	if consumer.Cancel != nil {
		consumer.Cancel()
	}

	if consumer.Reader != nil {
		consumer.Reader.Close()
	}

	consumer.IsActive = false
	conn.mu.Unlock()

	log.Printf("[Kafka] Consumer stopped: %s", consumerID)
	emitStreamMessage(app, connectionID, "system", "kafka", fmt.Sprintf("Consumer %s stopped", consumerID[:8]))

	return nil
}

func KafkaProduceMessage(app AppInterface, config ProducerConfig) error {
	log.Printf("[Kafka] Producing message to topic: %s", config.Topic)

	kafkaMutex.RLock()
	conn, exists := kafkaConnections[config.ConnectionID]
	kafkaMutex.RUnlock()

	if !exists {
		return fmt.Errorf("connection not found: %s", config.ConnectionID)
	}

	var compression kafka.Compression
	switch config.Compression {
	case "gzip":
		compression = kafka.Gzip
	case "snappy":
		compression = kafka.Snappy
	case "lz4":
		compression = kafka.Lz4
	case "zstd":
		compression = kafka.Zstd
	}

	writer := &kafka.Writer{
		Addr:         kafka.TCP(conn.Brokers...),
		Topic:        config.Topic,
		Balancer:     &kafka.LeastBytes{},
		MaxAttempts:  3,
		RequiredAcks: kafka.RequiredAcks(config.Acks),
		Async:        false,
		Compression:  compression,
	}
	defer writer.Close()

	msg := kafka.Message{
		Key:   []byte(config.Key),
		Value: []byte(config.Value),
		Time:  time.Now(),
	}

	if len(config.Headers) > 0 {
		headers := make([]kafka.Header, 0, len(config.Headers))
		for k, v := range config.Headers {
			headers = append(headers, kafka.Header{
				Key:   k,
				Value: []byte(v),
			})
		}
		msg.Headers = headers
	}

	if config.Partition >= 0 {
		msg.Partition = config.Partition
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := writer.WriteMessages(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to produce message: %w", err)
	}

	log.Printf("[Kafka] Message produced successfully to topic: %s", config.Topic)

	emitStreamMessage(app, config.ConnectionID, "outbound", "kafka", fmt.Sprintf("Key: %s\nValue: %s", config.Key, config.Value))

	return nil
}

func consumeMessages(app AppInterface, ctx context.Context, connectionID string, consumerID string, reader *kafka.Reader) {
	log.Printf("[Kafka] Consumer goroutine started: %s", consumerID)

	retryCount := 0
	maxRetries := 5

	for {
		select {
		case <-ctx.Done():
			log.Printf("[Kafka] Consumer goroutine stopped: %s", consumerID)
			return
		default:
			msg, err := reader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					// Context cancelled, exit gracefully
					return
				}

				// Handle coordinator not available error with retries
				if err.Error() != "" && retryCount < maxRetries {
					retryCount++
					log.Printf("[Kafka] Consumer %s waiting for coordinator (attempt %d/%d): %v",
						consumerID, retryCount, maxRetries, err)
					time.Sleep(time.Duration(retryCount) * time.Second)
					continue
				}

				log.Printf("[Kafka] Error reading message: %v", err)
				emitStreamMessage(app, connectionID, "error", "kafka", fmt.Sprintf("Error reading message: %v", err))
				time.Sleep(1 * time.Second) // Back off on errors
				continue
			}

			// Reset retry count on successful read
			retryCount = 0

			payload := fmt.Sprintf("Topic: %s\nPartition: %d\nOffset: %d\nKey: %s\nValue: %s\nTimestamp: %s",
				msg.Topic,
				msg.Partition,
				msg.Offset,
				string(msg.Key),
				string(msg.Value),
				msg.Time.Format(time.RFC3339),
			)

			if len(msg.Headers) > 0 {
				payload += "\nHeaders:"
				for _, h := range msg.Headers {
					payload += fmt.Sprintf("\n  %s: %s", h.Key, string(h.Value))
				}
			}

			emitStreamMessage(app, connectionID, "inbound", "kafka", payload)

			// NO COMMIT - we're not using consumer groups
		}
	}
}

func emitStreamMessage(app AppInterface, connectionID, direction, protocol, payload string) {
	app.EmitStreamMessage(connectionID, direction, protocol, payload)
}

func createSASLMechanism(mechanism, username, password string) (sasl.Mechanism, error) {
	switch mechanism {
	case "plain":
		return plain.Mechanism{
			Username: username,
			Password: password,
		}, nil
	case "scram-sha-256":
		return scram.Mechanism(scram.SHA256, username, password)
	case "scram-sha-512":
		return scram.Mechanism(scram.SHA512, username, password)
	default:
		return nil, fmt.Errorf("unsupported SASL mechanism: %s", mechanism)
	}
}
