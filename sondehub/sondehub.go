package sondehub

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)



type StreamConfig struct {
	Sondes           []string
	OnConnect        mqtt.OnConnectHandler
	OnMessage        func([]byte)
	OnDisconnect     mqtt.ConnectionLostHandler
	OnLog            mqtt.Logger
	AsJSON           bool
	AutoStartLoop     bool
	Prefix           string
}

type Stream struct {
	mqttc        mqtt.Client
	mqttMutex    sync.Mutex
	config       StreamConfig
	log          mqtt.Logger
	MessageHandler func([]byte) // Added field for message handler
}

type customLogger struct{}

func (l customLogger) Println(v ...interface{}) {
	fmt.Println(v...)
}

func NewStream(options ...func(*StreamConfig)) *Stream {
	config := StreamConfig{
		Sondes:       []string{"#"},
		AsJSON:       false,
		AutoStartLoop: true,
		Prefix:       "sondes",
	}

	for _, option := range options {
		option(&config)
	}

	s := &Stream{
		config: config,
		log:    config.OnLog,
	}
	s.WsConnect()

	return s
}

func WithSondes(sondes []string) func(*StreamConfig) {
	return func(c *StreamConfig) {
		c.Sondes = sondes
	}
}

func WithOnConnect(handler mqtt.OnConnectHandler) func(*StreamConfig) {
	return func(c *StreamConfig) {
		c.OnConnect = handler
	}
}

func WithOnMessage(handler func([]byte)) func(*StreamConfig) {
	return func(c *StreamConfig) {
		c.OnMessage = handler
	}
}

func WithOnDisconnect(handler mqtt.ConnectionLostHandler) func(*StreamConfig) {
	return func(c *StreamConfig) {
		c.OnDisconnect = handler
	}
}

func WithOnLog(logger mqtt.Logger) func(*StreamConfig) {
	return func(c *StreamConfig) {
		c.OnLog = logger
	}
}

func WithAsJSON(asJSON bool) func(*StreamConfig) {
	return func(c *StreamConfig) {
		c.AsJSON = asJSON
	}
}

func WithAutoStartLoop(autoStart bool) func(*StreamConfig) {
	return func(c *StreamConfig) {
		c.AutoStartLoop = autoStart
	}
}

func WithPrefix(prefix string) func(*StreamConfig) {
	return func(c *StreamConfig) {
		c.Prefix = prefix
	}
}


func (s *Stream) WsDisconnect() {
	s.mqttMutex.Lock()
	defer s.mqttMutex.Unlock()

	// Disconnect from Sondehub
	if s.mqttc != nil && s.mqttc.IsConnected() {
		s.mqttc.Disconnect(0)
	}
}



func (s *Stream) AddSonde(sonde string) {
	s.mqttMutex.Lock()
	defer s.mqttMutex.Unlock()

	if !s.ContainsSonde(sonde) {
		s.config.Sondes = append(s.config.Sondes, sonde)
		token := s.mqttc.Subscribe(fmt.Sprintf("%s/%s", s.config.Prefix, sonde), 0, nil)
		if token.Wait() && token.Error() != nil {
			fmt.Println("Error subscribing to topic:", token.Error())
			s.WsConnect()
		}
	}
}


func (s *Stream) RemoveSonde(sonde string) {
	s.mqttMutex.Lock()
	defer s.mqttMutex.Unlock()

	for i, v := range s.config.Sondes {
		if v == sonde {
			s.config.Sondes = append(s.config.Sondes[:i], s.config.Sondes[i+1:]...)
			token := s.mqttc.Unsubscribe(fmt.Sprintf("%s/%s", s.config.Prefix, sonde))
			if token.Wait() && token.Error() != nil {
				s.WsConnect()
			}
			break
		}
	}
}

func (s *Stream) onStreamMessage(client mqtt.Client, msg mqtt.Message) {
	if s.config.OnMessage != nil {
		payload := msg.Payload()
		// Removed fmt.Printf statement
		s.config.OnMessage(payload)
	}

	// Log the received message using the OnLog callback
	if s.log != nil {
		s.log.Println("Received message:", string(msg.Payload()))
	}
}

func (s *Stream) WsConnect() {
	s.mqttMutex.Lock()
	defer s.mqttMutex.Unlock()

	serverURL := s.GetURL()
	clientID := uuid.New().String()
	topic := "sondes/#"

	opts := mqtt.NewClientOptions()
	opts.AddBroker(serverURL)
	opts.SetClientID(clientID)

	// Set the message handler if provided
	if s.MessageHandler != nil {
		opts.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
			s.onStreamMessage(client, msg)
		})
	}

	s.mqttc = mqtt.NewClient(opts)

	if token := s.mqttc.Connect(); token.Wait() && token.Error() != nil {
		fmt.Println("Error connecting to MQTT:", token.Error())
		os.Exit(1)
	}

	if token := s.mqttc.Subscribe(topic, 0, nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}
}

func (s *Stream) GetURL() string {
	resp, err := http.Get("https://api.v2.sondehub.org/sondes/websocket")
	if err != nil {
		fmt.Println("Error getting URL:", err)
		return ""
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return ""
	}

	return string(data)
}


func (s *Stream) ContainsSonde(sonde string) bool {
	for _, v := range s.config.Sondes {
		if v == sonde {
			return true
		}
	}
	return false
}


