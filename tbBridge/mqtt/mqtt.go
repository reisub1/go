package mqtt

import (
	mqMessage "github.com/surgemq/message"
	// mq "github.com/surgemq/surgemq"
	mqService "github.com/surgemq/surgemq/service"
)

func Connect(host string, access_token string) (*mqService.Client, error) {
	c := &mqService.Client{}
	msgConnPacket := mqMessage.NewConnectMessage()
	msgConnPacket.SetVersion(4)
	msgConnPacket.SetCleanSession(true)
	msgConnPacket.SetClientId([]byte("gg"))
	msgConnPacket.SetKeepAlive(10)

	msgConnPacket.SetUsername([]byte(access_token))
	msgConnPacket.SetPassword([]byte("password"))

	err := c.Connect(host, msgConnPacket)
	return c, err
}

func Publish(c *mqService.Client, message string, topic string) {
	pubmsg := mqMessage.NewPublishMessage()
	pubmsg.SetTopic([]byte(topic))
	pubmsg.SetPayload([]byte(message))
	pubmsg.SetQoS(0)
	c.Publish(pubmsg, nil)
}
