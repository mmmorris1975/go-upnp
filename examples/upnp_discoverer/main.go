package main

import (
	"flag"
	"fmt"
	"github.com/mmmorris1975/go-upnp/discovery"
)

func main() {
	wait := flag.Duration("wait", discovery.DISCOVERY_WAIT_MAX_DURATION, "Duration for UPnP search request")
	target := flag.String("target", discovery.DISCOVERY_TARGET_DEFAULT, "UPnP target to search for")
	flag.Parse()

	r := discovery.NewSearchRequest()
	r.Target = *target
	r.Wait = *wait

	ch := make(chan *discovery.SearchResponse, 10)
	go discovery.Discover(r, ch)

	for r := range ch {
		fmt.Printf("%+v\n\n", r)
	}
}
