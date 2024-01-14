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

type Stream struct {
	mqttc     mqtt.Client
	mqttMutex sync.Mutex
	Sondes []string
	log             mqtt.Logger
	MessageHandler  func([]byte)
	stopLoopChannel chan struct{}
}

func NewStream(_MessageHandler func([]byte)) *Stream {

	s := &Stream{
		MessageHandler: _MessageHandler,
	}
	s.WsConnect()

	return s
}


func (s *Stream) RemoveSonde(sonde string) {
	s.mqttMutex.Lock()
	defer s.mqttMutex.Unlock()

	for i, v := range s.Sondes {
		if v == sonde {
			s.Sondes = append(s.Sondes[:i], s.Sondes[i+1:]...)
			token := s.mqttc.Unsubscribe(fmt.Sprintf("sondes/%s", sonde))
			if token.Wait() && token.Error() != nil {
				s.WsConnect()
			}
			break
		}
	}
}

func (s *Stream) ContainsSonde(sonde string) bool {
	s.mqttMutex.Lock()
	defer s.mqttMutex.Unlock()

	for _, v := range s.Sondes {
		if v == sonde {
			return true
		}
	}
	return false
}


func (s *Stream) onStreamMessage(client mqtt.Client, msg mqtt.Message) {
	if s.MessageHandler != nil {
		payload := msg.Payload()
		// Removed fmt.Printf statement
		s.MessageHandler(payload)
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





func (s *Stream) Disconnect() {
	s.mqttMutex.Lock()

	s.mqttc.Disconnect(0)
	fmt.Printf("Sondehub Disconnected\n")
}

