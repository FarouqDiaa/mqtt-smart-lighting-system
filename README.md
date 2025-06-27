# MQTT Smart Lighting System

A complete MQTT-based smart lighting control system built from scratch in Go, featuring a custom MQTT broker implementation and client applications for IoT device simulation.

## Description

This project implements a full-stack MQTT messaging system designed for smart home lighting control. The system consists of three main components:

- **Custom MQTT Broker**: A from-scratch implementation of an MQTT broker that handles client connections, subscriptions, and message routing
- **Smart Light Device**: An MQTT client that simulates a smart lamp pole, responding to lighting control commands
- **Control Interface**: A publisher client that sends random lighting commands to simulate user interactions

The project demonstrates real-world IoT communication patterns and provides a complete understanding of MQTT protocol implementation at the network level.

## Features

### Custom MQTT Broker
- **Full MQTT Protocol Support**: Handles CONNECT, PUBLISH, SUBSCRIBE, PING, and DISCONNECT messages
- **Multi-Client Management**: Concurrent handling of multiple MQTT clients
- **Topic-Based Routing**: Efficient message distribution to subscribed clients
- **Connection Management**: Automatic client cleanup and reconnection handling
- **Thread-Safe Operations**: Concurrent-safe client and subscription management

### Smart Device Simulation
- **Light State Management**: ON/OFF state tracking with duplicate command prevention
- **Brightness Control**: Variable dimming levels (0-100)
- **Real-time Response**: Immediate feedback to control commands
- **Connection Resilience**: Automatic reconnection and error handling

### Control Interface
- **Random Command Generation**: Simulates realistic user interactions
- **Multiple Command Types**: ON/OFF toggles and brightness adjustments
- **Continuous Operation**: Automated command sending with configurable intervals

## Architecture

```
┌─────────────────┐         MQTT Messages         ┌─────────────────┐
│   Controller    │ ──────────────────────────► │  Custom Broker  │
│   (Publisher)   │        Port 1883/TCP        │ (Message Router) │
└─────────────────┘                             └─────────────────┘
                                                          │
                                                          ▼
                                                ┌─────────────────┐
                                                │   Smart Light   │
                                                │  (Subscriber)   │
                                                └─────────────────┘
```

## Prerequisites

- Go 1.24+
- Network connectivity for MQTT communication

## Quick Start

1. **Clone the repository**
   ```bash
   git clone https://github.com/FarouqDiaa/mqtt-smart-lighting-system
   cd mqtt_smart_lighting
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Start the MQTT Broker**
   ```bash
   cd broker
   go run main.go
   ```

4. **Start the Smart Light Device** (in a new terminal)
   ```bash
   cd device
   go run main.go
   ```

5. **Start the Controller** (in a new terminal)
   ```bash
   cd controller
   go run main.go
   ```

## Project Structure

```
mqtt_smart_lighting/
├── Broker/                     # Custom MQTT broker implementation
│   ├── main.go                # Core broker with protocol handling
│   ├── go.mod
│   └── go.sum
├── LightLamp/                     # Smart light device (subscriber)
│   ├── main.go                # Light control logic
│   ├── go.mod
│   └── go.sum
├── Client/                 # Control interface (publisher)
│   ├── main.go                # Command generation and publishing
│   ├── go.mod
│   └── go.sum
```

## Configuration

### Default Settings
- **MQTT Port**: 1883 (standard MQTT port)
- **Broker Address**: localhost
- **Topic**: "LightControl"
- **QoS Level**: 0 (At most once delivery)
- **Command Interval**: 2 seconds

### Supported Commands
- `ON`: Turn light on
- `OFF`: Turn light off  
- `0-100`: Set brightness level (dimming)

## Sample Output

### Broker Output
```
MQTT Broker started on port 1883
Client LampPole connected
Client User connected
Client LampPole subscribed to topic: LightControl
Publishing to topic LightControl: ON
```

### Smart Light Device Output
```
Connected to MQTT Broker
Light turned ON
Light Dim changed to: 75
Light already ON
Light turned OFF
```

### Controller Output
```
Connected to MQTT Broker  
Published message: ON
Published message: 50
Published message: OFF
Published message: 100
```

## MQTT Protocol Implementation

The custom broker implements key MQTT packet types:

- **CONNECT (1)**: Client connection establishment
- **CONNACK (2)**: Connection acknowledgment
- **PUBLISH (3)**: Message publishing
- **SUBSCRIBE (8)**: Topic subscription
- **SUBACK (9)**: Subscription acknowledgment
- **PINGREQ (12)**: Keep-alive ping request
- **PINGRESP (13)**: Keep-alive ping response
- **DISCONNECT (14)**: Clean disconnection

## Technical Highlights

### Custom Protocol Implementation
- Raw TCP socket handling for MQTT packet processing
- Binary protocol parsing and generation
- Variable-length encoding for MQTT remaining length fields
- Proper MQTT packet structure with fixed and variable headers

### Concurrent Design
- Goroutine-based client handling for scalability
- Thread-safe data structures with mutex protection
- Non-blocking message delivery to prevent client blocking

### Error Handling
- Graceful client disconnection handling
- Connection loss detection and cleanup
- Malformed packet handling and recovery

## Development

### Running Individual Components

**Broker Only:**
```bash
cd broker && go run main.go
```

**Testing with External MQTT Client:**
```bash
# Subscribe to topic
mosquitto_sub -h localhost -p 1883 -t "LightControl"

# Publish message
mosquitto_pub -h localhost -p 1883 -t "LightControl" -m "ON"
```

## Technologies Used

- **Go 1.24**: Core implementation language
- **Eclipse Paho MQTT**: Client library for device implementations
- **Raw TCP Sockets**: Custom broker network handling
- **Goroutines**: Concurrent client management
- **Mutex Synchronization**: Thread-safe operations

## Use Cases

- **IoT Protocol Learning**: Understanding MQTT internals and implementation
- **Smart Home Development**: Foundation for lighting control systems
- **Message Broker Development**: Custom MQTT broker for specialized requirements
- **Network Programming**: TCP socket handling and binary protocol implementation
- **Concurrent Systems**: Multi-client server architecture patterns

## Future Enhancements

- QoS 1 and 2 support for guaranteed delivery
- Authentication and authorization mechanisms
- WebSocket support for web-based clients
- Message persistence and offline client support
- SSL/TLS encryption for secure communication

## Contributing

1. Fork the repository
2. Create a feature branch
3. Implement your changes
4. Add tests for new functionality
5. Submit a pull request

## License

This project was developed during an internship at IoTBlue.

---

*Developed during internship at IoTBlue*
