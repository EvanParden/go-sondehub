package main

import (
	"fmt"

	"github.com/EvanParden/go-sondehub/sondehub"
)

func onMessage(message []byte) {
	fmt.Println("msg")
}

func main() {
	sondehub.NewStream(
		sondehub.WithOnMessage(onMessage),
	)

	select {}
}
