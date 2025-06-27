package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

// MQTT Message Types
const (
	CONNECT     = 1
	CONNACK     = 2
	PUBLISH     = 3
	PUBACK      = 4
	SUBSCRIBE   = 8
	SUBACK      = 10
	UNSUBSCRIBE = 10
	UNSUBACK    = 11
	PINGREQ     = 12
	PINGRESP    = 13
	DISCONNECT  = 14
)

type Client struct {
	conn     net.Conn
	clientID string
	topics   map[string]bool
	mu       sync.RWMutex
}

type Message struct {
	topic   string
	payload []byte
	qos     byte
}

type Broker struct {
	clients     map[string]*Client
	clientsMu   sync.RWMutex
	subscribers map[string][]*Client
	subsMu      sync.RWMutex
}

func NewBroker() *Broker {
	return &Broker{
		clients:     make(map[string]*Client),
		subscribers: make(map[string][]*Client),
	}
}

func (b *Broker) addClient(client *Client) {
	b.clientsMu.Lock()
	defer b.clientsMu.Unlock()
	b.clients[client.clientID] = client
	log.Printf("Client %s connected", client.clientID)
}

func (b *Broker) removeClient(clientID string) {
	b.clientsMu.Lock()
	defer b.clientsMu.Unlock()

	if client, exists := b.clients[clientID]; exists {
		// Remove from all subscriptions
		b.subsMu.Lock()
		for topic := range client.topics {
			b.removeSubscription(topic, client)
		}
		b.subsMu.Unlock()

		client.conn.Close()
		delete(b.clients, clientID)
		log.Printf("Client %s disconnected", clientID)
	}
}

func (b *Broker) subscribe(topic string, client *Client) {
	b.subsMu.Lock()
	defer b.subsMu.Unlock()

	b.subscribers[topic] = append(b.subscribers[topic], client)
	client.mu.Lock()
	client.topics[topic] = true
	client.mu.Unlock()

	log.Printf("Client %s subscribed to topic: %s", client.clientID, topic)
}

func (b *Broker) removeSubscription(topic string, client *Client) {
	subscribers := b.subscribers[topic]
	for i, sub := range subscribers {
		if sub == client {
			b.subscribers[topic] = append(subscribers[:i], subscribers[i+1:]...)
			break
		}
	}
	if len(b.subscribers[topic]) == 0 {
		delete(b.subscribers, topic)
	}
}

func (b *Broker) publish(topic string, payload []byte) {
	b.subsMu.RLock()
	subscribers := b.subscribers[topic]
	b.subsMu.RUnlock()

	if len(subscribers) == 0 {
		log.Printf("No subscribers for topic: %s", topic)
		return
	}

	log.Printf("Publishing to topic %s: %s", topic, string(payload))

	for _, client := range subscribers {
		go func(c *Client) {
			if err := c.sendPublish(topic, payload); err != nil {
				log.Printf("Failed to send message to client %s: %v", c.clientID, err)
				b.removeClient(c.clientID)
			}
		}(client)
	}
}

func (c *Client) sendPublish(topic string, payload []byte) error {
	// PUBLISH packet format:
	// Fixed header: [PUBLISH(3<<4) | flags, remaining length]
	// Variable header: [topic length MSB, topic length LSB, topic string]
	// Payload: [message payload]

	topicBytes := []byte(topic)
	topicLen := len(topicBytes)
	payloadLen := len(payload)

	// Calculate remaining length
	remainingLen := 2 + topicLen + payloadLen

	packet := make([]byte, 0, 2+remainingLen)

	// Fixed header
	packet = append(packet, 0x30) // PUBLISH packet type
	packet = append(packet, byte(remainingLen))

	// Variable header - topic
	packet = append(packet, byte(topicLen>>8), byte(topicLen&0xFF))
	packet = append(packet, topicBytes...)

	// Payload
	packet = append(packet, payload...)

	_, err := c.conn.Write(packet)
	return err
}

func (c *Client) handleConnection(broker *Broker) {
	defer broker.removeClient(c.clientID)

	reader := bufio.NewReader(c.conn)
	for {
		// Read fixed header
		firstByte, err := reader.ReadByte()
		if err != nil {
			log.Printf("Error reading from client %s: %v", c.clientID, err)
			return
		}

		msgType := (firstByte >> 4) & 0x0F

		// Read remaining length
		remainingLen, err := readRemainingLength(reader)
		if err != nil {
			log.Printf("Error reading remaining length: %v", err)
			return
		}

		// Read the rest of the packet
		packet := make([]byte, remainingLen)
		if remainingLen > 0 {
			_, err = reader.Read(packet)
			if err != nil {
				log.Printf("Error reading packet: %v", err)
				return
			}
		}
		// if msgType != 12 {
		// 	log.Printf("First Byte: %v", firstByte)
		// 	log.Printf("Message type: %v", msgType)
		// 	log.Printf("Packet: %v", packet)
		// }
		switch msgType {
		case CONNECT:
			c.handleConnect(packet)
		case PUBLISH:
			c.handlePublish(packet, broker)
		case SUBSCRIBE:
			c.handleSubscribe(packet, broker)
		case PINGREQ:
			c.handlePingReq()
		case DISCONNECT:
			log.Printf("Client %s disconnecting", c.clientID)
			return
		}
	}
}

func (c *Client) handleConnect(packet []byte) {
	// Send CONNACK
	connack := []byte{0x20, 0x02, 0x00, 0x00} // CONNACK with success
	c.conn.Write(connack)
	log.Printf("Sent CONNACK to client %s", c.clientID)
}

func (c *Client) handlePublish(packet []byte, broker *Broker) {
	if len(packet) < 2 {
		return
	}

	// Extract topic
	topicLen := int(packet[0])<<8 | int(packet[1])
	if len(packet) < 2+topicLen {
		return
	}

	topic := string(packet[2 : 2+topicLen])
	payload := packet[2+topicLen:]

	broker.publish(topic, payload)
}

func (c *Client) handleSubscribe(packet []byte, broker *Broker) {
	log.Printf("Handling Subscribe")
	if len(packet) < 4 {
		return
	}

	// Skip packet identifier (2 bytes)
	pos := 2

	for pos < len(packet) {
		if pos+2 >= len(packet) {
			break
		}

		// Read topic length
		topicLen := int(packet[pos])<<8 | int(packet[pos+1])
		pos += 2

		if pos+topicLen >= len(packet) {
			break
		}

		// Read topic
		topic := string(packet[pos : pos+topicLen])
		pos += topicLen

		// Skip QoS byte
		pos++

		broker.subscribe(topic, c)
	}

	// Send SUBACK
	suback := []byte{0x90, 0x03, packet[0], packet[1], 0x00} // SUBACK with granted QoS 0
	c.conn.Write(suback)
}

func (c *Client) handlePingReq() {
	// Send PINGRESP
	pingresp := []byte{0xD0, 0x00}
	c.conn.Write(pingresp)
}

func readRemainingLength(reader *bufio.Reader) (int, error) {
	var length int
	var shift uint

	for {
		if shift >= 28 {
			return 0, fmt.Errorf("remaining length exceeds maximum")
		}

		b, err := reader.ReadByte()
		if err != nil {
			return 0, err
		}

		length |= int(b&0x7F) << shift
		if (b & 0x80) == 0 {
			break
		}
		shift += 7
	}

	return length, nil
}

func main() {
	broker := NewBroker()

	listener, err := net.Listen("tcp", ":1883")
	if err != nil {
		log.Fatal("Failed to start broker:", err)
	}
	defer listener.Close()

	log.Println("MQTT Broker started on port 1883")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Failed to accept connection:", err)
			continue
		}

		client := &Client{
			conn:     conn,
			clientID: fmt.Sprintf("client_%d", time.Now().Unix()),
			topics:   make(map[string]bool),
		}

		broker.addClient(client)
		go client.handleConnection(broker)
	}
}
