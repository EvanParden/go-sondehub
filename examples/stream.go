package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/EvanParden/go-sondehub/sondehub"
)

func onMessage(message []byte) {
	fmt.Println(string(message))
}

func main() {
	stream := sondehub.NewStream(
		[]string{"#"},
		nil,
		onMessage,
		nil,
		nil,
		false,
		"sondes",
	)

	// Use a channel to wait for an interrupt signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Keep the program running until an interrupt signal is received
	select {
	case <-interrupt:
		fmt.Println("Received interrupt signal. Exiting...")
		// Gracefully close your MQTT connection or perform any cleanup if needed
		print(stream)
	}
}
