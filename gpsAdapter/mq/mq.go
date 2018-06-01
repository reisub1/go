package mq

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	// mq "github.com/surgemq/surgemq"
	// mqService "github.com/surgemq/surgemq/service"
)

// Connects to the given broker with the username set as the accesstoken for ThingsBoard
// Returns a client object which must be stored to reuse this connection
func Connect(broker string, access_token string) (*mqtt.Client, error) {
	var e error

	// Set up the client parameters
	opts := mqtt.NewClientOptions().AddBroker(broker)
	opts.SetUsername(access_token)
	opts.SetClientID("tbBridge")
	opts.SetKeepAlive(0)
	client := mqtt.NewClient(opts)

	// Attempt a connection
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		e = token.Error()
	}

	// Return the client for further calls
	return &client, e
}

// A simple wrapper around Paho.Mqtt.Golang that publishes the message to the topic, given the client
func Publish(c *mqtt.Client, message string, topic string) error {
	var e error
	if token := (*c).Publish(topic, 0, false, message); token.Wait() && token.Error() != nil {
		e = token.Error()
	}
	return e
}

type Client mqtt.Client
