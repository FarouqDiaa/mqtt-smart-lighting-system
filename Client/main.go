package main

import (
	"fmt"
	"math/rand"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	broker   = "tcp://localhost:1883"
	clientID = "User"
	topic    = "LightControl"
)

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected to MQTT Broker")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connection lost: %v", err)
}

func main() {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(clientID)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	for {
		message := generateRandomMessage()
		token := client.Publish(topic, 0, false, message)
		token.Wait()
		fmt.Printf("Published message: %s\n", message)
		time.Sleep(2 * time.Second)
	}
}

func generateRandomMessage() string {
	messages := []string{
		"ON",
		"OFF",
		"50",
		"100",
		"0",
	}
	return messages[rand.Intn(len(messages))]
}

func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
}
