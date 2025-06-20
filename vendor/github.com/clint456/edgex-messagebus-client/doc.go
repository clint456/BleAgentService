/*
Package messagebus provides a high-level client library for EdgeX Foundry MessageBus operations.

# Overview

This package simplifies interaction with EdgeX Foundry's MessageBus system, offering
a clean and intuitive API for message publishing, subscribing, and request-response
patterns. It supports both MQTT and NATS protocols and provides comprehensive
connection management, error handling, and logging capabilities.

# Key Features

- Complete EdgeX MessageBus support based on official go-mod-messaging library
- Automatic connection management with reconnection capabilities
- Thread-safe operations with proper synchronization
- Support for multiple data types (JSON, binary, string)
- Request-response pattern implementation
- Comprehensive error handling and logging
- Health checking capabilities
- Subscription management with custom message handlers

# Quick Start

	package main

	import (
		"log"
		"time"

		messagebus "github.com/clint456/edgex-messagebus-client"
		"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
		"github.com/edgexfoundry/go-mod-messaging/v4/pkg/types"
	)

	func main() {
		// Create logger
		lc := logger.NewClient("MyApp", "INFO")

		// Configure client
		config := messagebus.Config{
			Host:     "localhost",
			Port:     1883,
			Protocol: "tcp",
			Type:     "mqtt",
			ClientID: "my-client",
		}

		// Create and connect client
		client, err := messagebus.NewClient(config, lc)
		if err != nil {
			log.Fatal(err)
		}

		if err := client.Connect(); err != nil {
			log.Fatal(err)
		}
		defer client.Disconnect()

		// Publish message
		data := map[string]interface{}{
			"temperature": 25.6,
			"timestamp":   time.Now(),
		}
		client.Publish("sensors/temperature", data)

		// Subscribe to messages with wildcard support
		handler := func(topic string, message types.MessageEnvelope) error {
			// topic parameter contains the actual received topic path
			log.Printf("Received message on %s: %s", topic, string(message.Payload.([]byte)))
			return nil
		}
		client.SubscribeSingle("sensors/#", handler)

		// Keep running
		time.Sleep(10 * time.Second)
	}

# Configuration

The Config struct provides all necessary configuration options:

	type Config struct {
		Host     string  // MessageBus broker host
		Port     int     // MessageBus broker port
		Protocol string  // Connection protocol (tcp, ssl, ws, wss)
		Type     string  // MessageBus type (mqtt, nats)
		ClientID string  // Unique client identifier
		Username string  // Authentication username (optional)
		Password string  // Authentication password (optional)
		QoS      int     // Quality of Service level (0, 1, 2)
	}

# Error Handling

The package provides comprehensive error handling through multiple mechanisms:

1. Return values: Most methods return an error that should be checked
2. Error channel: Subscribe to the error channel for asynchronous error notifications
3. Logging: All operations are logged through the provided LoggingClient

Example error handling:

	// Check return errors
	if err := client.Connect(); err != nil {
		log.Printf("Connection failed: %v", err)
	}

	// Monitor error channel
	go func() {
		for err := range client.GetErrorChannel() {
			log.Printf("MessageBus error: %v", err)
		}
	}()

# Thread Safety

All client operations are thread-safe and can be called concurrently from multiple
goroutines. The package uses appropriate synchronization mechanisms to ensure data
consistency and prevent race conditions.

# Best Practices

1. Always check connection status before performing operations
2. Use defer client.Disconnect() to ensure proper cleanup
3. Monitor the error channel for asynchronous errors
4. Use appropriate QoS levels based on your reliability requirements
5. Implement proper message handlers that return errors for failed processing
6. Use correlation IDs for request-response patterns
7. Perform regular health checks in long-running applications

# Dependencies

This package depends on the official EdgeX Foundry libraries:
- github.com/edgexfoundry/go-mod-core-contracts/v4
- github.com/edgexfoundry/go-mod-messaging/v4

# License

This package is licensed under the Apache License 2.0.
*/
package messagebus
