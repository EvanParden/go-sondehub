# SondeHub Go Package

## Example Streaming live telemetry data
```go
package main

import (
	"fmt"
	"github.com/EvanParden/go-sondehub/sondehub"
)

func onMessage(message []byte) {
	fmt.Println(string(message))
}

func main() {
	sondehub.NewStream(onMessage)

	select {}
}

```


## functions

### `NewStream(_MessageHandler func([]byte)) *Stream`

Creates a new SondeHub stream with the specified message handler.

- Parameters:
  - `_MessageHandler`: A function that handles incoming MQTT messages.

### `AddSonde(sonde string)`

Adds a sonde to the list of subscribed sondes.

- Parameters:
  - `sonde`: The name of the sonde to add.

### `RemoveSonde(sonde string)`

Removes a sonde from the list of subscribed sondes.

- Parameters:
  - `sonde`: The name of the sonde to remove.

### `ContainsSonde(sonde string) bool`

Checks if a specific sonde is present in the list of subscribed sondes.

- Parameters:
  - `sonde`: The name of the sonde to check.

- Returns:
  - `true` if the sonde is present, `false` otherwise.

### `Disconnect()`

Disconnects the SondeHub stream from the MQTT server.

---

### Example Usage:

```go
package main

import (
	"fmt"
	"time"
	"github.com/EvanParden/go-sondehub/sondehub"
)

func main() {
    // Create a new Stream with a message handler
    stream := sondehub.NewStream(onMessage)

    // Add sondes
    stream.AddSonde("sonde1")
    stream.AddSonde("sonde2")

    // Check if a sonde is present
    if stream.ContainsSonde("sonde1") {
        fmt.Println("sonde1 is present")
    } else {
        fmt.Println("sonde1 is not present")
    }

    // Remove a sonde
    stream.RemoveSonde("sonde1")

    // Check again after removal
    if stream.ContainsSonde("sonde1") {
        fmt.Println("sonde1 is present")
    } else {
        fmt.Println("sonde1 is not present")
    }

    // Keep the program running
    time.Sleep(30 * time.Second)
}

func onMessage(message []byte) {
    fmt.Println("Received message:", string(message))
}
