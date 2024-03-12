package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/EvanParden/go-sondehub/sondehub"
	"github.com/gorilla/websocket"
)

var (
    upgrader     = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
    clients      = make(map[*websocket.Conn]bool)
    activeClients int // Counter for active WebSocket clients
    clientsMutex  sync.Mutex
    sonde         *sondehub.Stream
    sondeMutex    sync.Mutex
    stopSonde     chan struct{}
)

func handleWebSocketConnection(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println(err)
        return
    }
    defer conn.Close()

    // Register the new client
    clientsMutex.Lock()
    clients[conn] = true
    activeClients++
    clientsMutex.Unlock()

    // Start sondehub if it's the first client
    sondeMutex.Lock()
    if sonde == nil {
        sonde = sondehub.NewStream(onMessage)

        stopSonde = make(chan struct{})
        go runSondeHub(stopSonde)
    }
    sondeMutex.Unlock()

    for {
        _, _, err := conn.ReadMessage()
        if err != nil {
            // Remove the client if there's an error (e.g., disconnected)
            clientsMutex.Lock()
            delete(clients, conn)
            activeClients--
            clientsMutex.Unlock()

            // Disconnect from sondehub if there are no active clients
            sondeMutex.Lock()
            if activeClients == 0 && sonde != nil {
                close(stopSonde) // Signal to stop sondehub goroutine
                sonde.Disconnect()
                sonde = nil
            }
            sondeMutex.Unlock()
            break
        }
    }
}




func runSondeHub(stopCh <-chan struct{}) {
	for {
		select {
		case <-stopCh:
			sondeMutex.Lock()
			if sonde != nil {
				sonde.Disconnect()

				sonde = nil
			}
			sondeMutex.Unlock()
			return // Stop the sondehub goroutine
		default:
			// Your sondehub logic goes here
			time.Sleep(1 * time.Second)
		}
	}
}

func onMessage(message []byte) {
	// Broadcast the message to all connected WebSocket clients
	clientsMutex.Lock()
	defer clientsMutex.Unlock()
	fmt.Print("msg \n")

	for client := range clients {
		err := client.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			// Remove the client if there's an error (e.g., disconnected)
			delete(clients, client)
		}
	}
}

func main() {
	// Start the WebSocket server
	http.HandleFunc("/ws", handleWebSocketConnection)
	go func() {
		log.Fatal(http.ListenAndServe(":8070", nil))
	}()

	select {}
}
