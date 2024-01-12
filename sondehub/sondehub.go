package sondehub

import (
	"fmt"
	"net/http"
	"os"
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
)

const (
	S3Bucket = "sondehub-history"
)

type Stream struct {
	mqttc         mqtt.Client
	sondes        []string
	asJSON        bool
	onConnect     mqtt.OnConnectHandler
	onMessage     func([]byte)
	onDisconnect  mqtt.ConnectionLostHandler
	onLog         mqtt.Logger
	prefix        string
	mqttMutex     sync.Mutex
}

type customLogger struct{}

func (l customLogger) Println(v ...interface{}) {
	fmt.Println(v...)
}

func NewStream(sondes []string, onConnect mqtt.OnConnectHandler, onMessage func([]byte), onDisconnect mqtt.ConnectionLostHandler, onLog mqtt.Logger, asJSON bool, prefix string) *Stream {
	s := &Stream{
		sondes:        sondes,
		asJSON:        asJSON,
		onConnect:     onConnect,
		onMessage:     onMessage,
		onDisconnect:  onDisconnect,
		onLog:         onLog,
		prefix:        prefix,
	}

	// Initialize mqttc before calling wsConnect
	opts := mqtt.NewClientOptions()
	opts.SetClientID(uuid.New().String()) // Set a UUID as the client ID
	s.mqttc = mqtt.NewClient(opts)
	s.wsConnect()

	return s
}

func (s *Stream) AddSonde(sonde string) {
    s.mqttMutex.Lock()
    defer s.mqttMutex.Unlock()

    if !s.containsSonde(sonde) {
        s.sondes = append(s.sondes, sonde)
        token := s.mqttc.Subscribe(fmt.Sprintf("%s/%s", s.prefix, sonde), 0, nil)
        if token.Wait() && token.Error() != nil {
            fmt.Println("Error subscribing to topic:", token.Error())
            s.wsConnect()
        }
    }
}

func (s *Stream) RemoveSonde(sonde string) {
	s.mqttMutex.Lock()
	defer s.mqttMutex.Unlock()

	for i, v := range s.sondes {
		if v == sonde {
			s.sondes = append(s.sondes[:i], s.sondes[i+1:]...)
			token := s.mqttc.Unsubscribe(fmt.Sprintf("%s/%s", s.prefix, sonde))
			if token.Wait() && token.Error() != nil {
				s.wsConnect()
			}
			break
		}
	}
}

func (s *Stream) onStreamMessage(client mqtt.Client, msg mqtt.Message) {
    if s.onMessage != nil {
        payload := msg.Payload()
        fmt.Printf("Received message: %s\n", payload)
        s.onMessage(payload)
    }
}

func (s *Stream) wsConnect() {

	s.mqttMutex.Lock()
    defer s.mqttMutex.Unlock()

    // resURL := s.getURL()
    // urlParts, err := url.Parse(resURL)
    // if err != nil {
    //     fmt.Println("Error parsing URL:", err)
    //     return
    // }

	// fmt.Print(resURL)
	// fmt.Print(urlParts.RawPath)

    // headers := map[string][]string{
    //     "Host": {urlParts.Host},
    // }

    s.mqttc = mqtt.NewClient(mqtt.NewClientOptions().
        AddBroker("wss://ws-reader.v2.sondehub.org").
        SetClientID("testhkkhgjgyguid").
        SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
            s.onStreamMessage(client, msg)
        }).
        SetOnConnectHandler(s.onConnect).
        SetConnectionLostHandler(s.onDisconnect))

    if token := s.mqttc.Connect(); token.Wait() && token.Error() != nil {
        fmt.Println("Error connecting to MQTT:", token.Error())
        os.Exit(1)
    }






}

func (s *Stream) getURL() string {
	resp, err := http.Get("https://api.v2.sondehub.org/sondes/websocket")
	if err != nil {
		fmt.Println("Error getting URL:", err)
		return ""
	}
	defer resp.Body.Close()

	var data []byte
	_, err = resp.Body.Read(data)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return ""
	}

	return string(data)
}

func (s *Stream) containsSonde(sonde string) bool {
	for _, v := range s.sondes {
		if v == sonde {
			return true
		}
	}
	return false
}


