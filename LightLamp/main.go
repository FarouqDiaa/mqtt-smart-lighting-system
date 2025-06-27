package main

import (
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var lightState = "ON"

const (
	broker   = "tcp://localhost:1883"
	clientID = "LampPole"
	topic    = "LightControl"
)

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected to MQTT Broker")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connection lost: %v", err)
}
var msgHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	var messageContent = string(msg.Payload())
	if messageContent == "ON" || messageContent == "OFF" {
		if lightState == "ON" && messageContent == "ON" || lightState == "OFF" && messageContent == "OFF" {
			fmt.Printf("Light already %v", messageContent)
			fmt.Println()
			return
		}
		lightState = messageContent
		fmt.Printf("Light turned %v", messageContent)
		fmt.Println()
	} else {
		fmt.Printf("Light Dim changed to: %v", messageContent)
		fmt.Println()
	}
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
	token := client.Subscribe(topic, 0, msgHandler)
	for {
		token.Wait()
	}
}
