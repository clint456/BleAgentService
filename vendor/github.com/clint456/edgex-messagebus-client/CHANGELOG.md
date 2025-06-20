# Changelog

All notable changes to this project will be documented in this file.

## [v1.1.0] - 2024-12-19

### Added
- **Enhanced Wildcard Subscription Support**: Message handlers now receive the actual topic path instead of the wildcard pattern
- **Automatic Topic Resolution**: When subscribing to wildcard topics (e.g., `edgex/events/#`), the handler function receives the specific topic path (e.g., `edgex/events/device/sensor01`)
- **New Wildcard Example**: Added comprehensive example demonstrating wildcard subscription and topic-specific message handling
- **Improved Message Processing**: Enhanced message handling to utilize `ReceivedTopic` field from MessageEnvelope

### Enhanced
- **Message Handler Interface**: The `MessageHandler` function now receives the actual received topic path, enabling topic-specific processing logic
- **Documentation Updates**: Updated README.md, USAGE.md, and doc.go with wildcard subscription examples and best practices
- **Example Improvements**: Enhanced main example to demonstrate multiple topic publishing and specific topic reception

### Technical Details
- Modified `handleMessages` function in `client.go` to use `message.ReceivedTopic` when available
- Maintains backward compatibility by falling back to subscription topic if `ReceivedTopic` is empty
- Added comprehensive wildcard subscription example in `example/wildcard/main.go`

## [v1.1.0] - 2024-12-19

### Added
- **Enhanced Wildcard Subscription Support**: Message handlers now receive the actual topic path instead of the wildcard pattern
- **Automatic Topic Resolution**: When subscribing to wildcard topics (e.g., `edgex/events/#`), the handler function receives the specific topic path (e.g., `edgex/events/device/sensor01`)
- **New Wildcard Example**: Added comprehensive example demonstrating wildcard subscription and topic-specific message handling
- **Improved Message Processing**: Enhanced message handling to utilize `ReceivedTopic` field from MessageEnvelope

### Enhanced
- **Message Handler Interface**: The `MessageHandler` function now receives the actual received topic path, enabling topic-specific processing logic
- **Documentation Updates**: Updated README.md, USAGE.md, and doc.go with wildcard subscription examples and best practices
- **Example Improvements**: Enhanced main example to demonstrate multiple topic publishing and specific topic reception

### Technical Details
- Modified `handleMessages` function in `client.go` to use `message.ReceivedTopic` when available
- Maintains backward compatibility by falling back to subscription topic if `ReceivedTopic` is empty
- Added comprehensive wildcard subscription example in `example/wildcard/main.go`

## [v0.1.0] - 2024-12-19

### Added
- Initial release of EdgeX MessageBus Client module
- Complete EdgeX MessageBus client implementation
- Support for MQTT publish/subscribe operations
- Thread-safe operations with proper error handling
- Request-response pattern support
- Binary data publishing and subscribing
- Health check functionality
- Client information retrieval
- Comprehensive error handling and logging
- Full documentation with examples
- Apache 2.0 license
