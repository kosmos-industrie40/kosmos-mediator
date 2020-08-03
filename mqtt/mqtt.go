package mqtt

import (
	"crypto/tls"
	"fmt"
	"math/rand"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"k8s.io/klog"
)

// Mqtt representing the mqtt connection
type Mqtt struct {
	client      MQTT.Client
	receiveChan chan<- Message
	sendChan    <-chan Message
}

// Message is the struct which representing the current message
type Message struct {
	Topic   string
	Payload []byte
}

func Connect(receiveChan chan<- Message, sendChan <-chan Message, clientID string, server string, username string, password string, topic string) (Mqtt, error) {
	var tlsConfig *tls.Config
	tlsConfig = &tls.Config{InsecureSkipVerify: true, ClientAuth: tls.NoClientCert}
	opts := MQTT.NewClientOptions().AddBroker(server).SetClientID(fmt.Sprintf("%s-%d", clientID, rand.Int31())).SetCleanSession(true)
	opts.SetAutoReconnect(true)
	if username != "" {
		klog.Infof("found username: %s", username)
		opts.SetUsername(username)
		if password != "" {
			opts.SetPassword(password)
		}
	}

	opts.SetTLSConfig(tlsConfig)
	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return Mqtt{}, token.Error()
	}

	mqtt := Mqtt{
		receiveChan: receiveChan,
		sendChan:    sendChan,
		client:      client,
	}

	return mqtt, nil
}

func (m Mqtt) send() {
	for {
		msg := <-m.sendChan
		token := m.client.Publish(msg.Topic, 1, false, msg.Payload)
		if token.Wait() && token.Error() != nil {
			klog.Errorf("could not send message; mqtt error: %s\n", token.Error())
		}
	}
}

func (m Mqtt) handle(client MQTT.Client, msg MQTT.Message) {
	klog.Infof("handle incomming request of topic: %s", msg.Topic())
	message := Message{Topic: msg.Topic(), Payload: msg.Payload()}
	m.receiveChan <- message
}

func (m Mqtt) receive(topic string) {
	token := m.client.Subscribe(topic, 1, m.handle)
	if token.Wait() && token.Error() != nil {
		klog.Errorf("Error by receiving message. MQTT error: %s\n", token.Error())
	}
}
