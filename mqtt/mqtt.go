package mqttClient

import (
	"crypto/tls"
	"fmt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"math/rand"
	"time"
)

type MqttWrapper struct {
	clientID         string
	client           MQTT.Client
	subscribedTopics []string
	Verbose          bool
	recMsg           chan MQTT.Message
}

type Msg struct {
	Topic string
	Msg   []byte
}

var defaultPublishHandler MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
	//implements what happens with received messages subscribed to before
	fmt.Printf("Rec at DefaultHandler: TOPIC: %s\n", msg.Topic(), msg.Topic())
}

func (m *MqttWrapper) Init(username, password, host string, port int, tls bool) error {
	mq := *m
	rand.Seed(time.Now().UnixNano())
	mq.clientID = fmt.Sprintf("connector-%d", rand.Int31())
	err := m.connect(host, m.clientID, username, password, port, tls)
	if err != nil {
		return err
	}
	return nil
}

func (m *MqttWrapper) connect(host, deviceId, user, password string, port int, tlsVerify bool) error {

	clientOpts := MQTT.NewClientOptions().AddBroker(fmt.Sprintf("tcp://%s:%d", host, port)).SetClientID(deviceId).SetCleanSession(true)
	//clientOpts.SetDefaultPublishHandler(defaultPublishHandler)

	if user != "" {
		clientOpts.SetUsername(user)
		if password != "" {
			clientOpts.SetPassword(password)
		}
	}

	if tlsVerify {
		tlsConfig := &tls.Config{ClientAuth: tls.NoClientCert}
		clientOpts.SetTLSConfig(tlsConfig)
	} else {
		tlsConfig := &tls.Config{InsecureSkipVerify: true, ClientAuth: tls.NoClientCert}
		clientOpts.SetTLSConfig(tlsConfig)
	}

	m.client = MQTT.NewClient(clientOpts)

	if tokenClient := m.client.Connect(); tokenClient.Wait() && tokenClient.Error() != nil {
		return tokenClient.Error()
	}

	return nil
}

func (m *MqttWrapper) Disconnect(host, deviceId, user, password string, port int, tlsVerify bool) error {
	for _, topic := range m.subscribedTopics {
		if token := m.client.Unsubscribe(topic); token.Wait() && token.Error() != nil {
			fmt.Println(token.Error())
			return token.Error()
		}
	}
	m.client.Disconnect(250)
	return nil
}

func (m *MqttWrapper) Subscribe(topic string, callBack MQTT.MessageHandler) error {
	if token := m.client.Subscribe(topic, 2, callBack); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		return token.Error()
	}
	m.subscribedTopics = append(m.subscribedTopics, topic)

	if m.Verbose == true {
		fmt.Printf("Subscribed to topic %s \n", topic)
	}
	return nil
}

func (m *MqttWrapper) Publish(topic string, msg []byte) error {
	token := m.client.Publish(topic, 2, false, msg)
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}

	if m.Verbose == true {
		fmt.Printf("Pub: TOPIC %s MSG: %s \n", topic, msg)
	}

	return nil
}
