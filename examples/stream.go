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
