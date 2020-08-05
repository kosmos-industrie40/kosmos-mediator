package mqttClient

import (
	"crypto/tls"
	"fmt"
	"math/rand"
	"os"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"k8s.io/klog"
)

// MqttWrapper representing the mqtt connection in this programm
type MqttWrapper struct {
	clientID         string
	client           MQTT.Client
	subscribedTopics []string
}

// Msg representing an message which can be used to publish to an MQTT broker
type Msg struct {
	Topic string
	Msg   []byte
}

// Init initialise the mqtt wrapper
func (m *MqttWrapper) Init(username, password, host string, port int, tls bool) error {
	mq := *m
	rand.Seed(time.Now().UnixNano())

	hostname, err := os.Hostname()
	if err != nil {
		klog.Errorf("could not query hostname; err")
		os.Exit(1)
	}

	mq.clientID = fmt.Sprintf("%s-connector-%d", hostname, rand.Int31())
	err = m.connect(host, m.clientID, username, password, port, tls)
	if err != nil {
		return err
	}
	return nil
}

func (m *MqttWrapper) connect(host, deviceId, user, password string, port int, tlsVerify bool) error {

	clientOpts := MQTT.NewClientOptions().AddBroker(fmt.Sprintf("tcp://%s:%d", host, port)).SetClientID(deviceId).SetCleanSession(true)

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

// Disconnect will disconnect the connection to the mqtt broker
func (m *MqttWrapper) Disconnect(host, deviceId, user, password string, port int, tlsVerify bool) error {
	for _, topic := range m.subscribedTopics {
		if token := m.client.Unsubscribe(topic); token.Wait() && token.Error() != nil {
			return token.Error()
		}
	}
	m.client.Disconnect(250)
	return nil
}

// Subscribe subscribe to a mqtt topic and set the handler function of this topic
func (m *MqttWrapper) Subscribe(topic string, callBack MQTT.MessageHandler) error {
	if token := m.client.Subscribe(topic, 1, callBack); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	m.subscribedTopics = append(m.subscribedTopics, topic)

	klog.V(2).Infof("Subscribed to topic %s", topic, " \n")
	return nil
}

// Publish publish an mqtt message to the mqtt broker
func (m *MqttWrapper) Publish(msg Msg) error {
	token := m.client.Publish(msg.Topic, 1, false, msg.Msg)
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}

	klog.V(2).Infof("Pub: TOPIC %s  MSG: %s\n", msg.Topic, string(msg.Msg))

	return nil
}
