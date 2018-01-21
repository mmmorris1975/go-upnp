package main

import (
	"github.com/mmmorris1975/upnp/eventing"
	"log"
)

func main() {
	ch := make(chan *eventing.Event, 10)
	go eventing.ListenMulticastEvents(ch)

	for e := range ch {
		log.Printf("EVENT: %v", e)
	}
}
