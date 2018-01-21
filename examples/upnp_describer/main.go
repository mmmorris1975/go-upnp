package main

import (
	"flag"
	"github.com/mmmorris1975/upnp/description"
	"github.com/mmmorris1975/upnp/discovery"
	"log"
)

func main() {
	var desc interface{}
	var err error

	wait := flag.Duration("wait", discovery.DISCOVERY_WAIT_MAX_DURATION, "Duration for UPnP search request")
	svc := flag.String("service", "", "UPnP service to describe")
	dev := flag.String("device", "", "UPnP device to describe")
	flag.Parse()

	switch {
	case len(*svc) > 0:
		desc, err = description.DiscoverServiceDescription(*svc, *wait)
		if err != nil {
			log.Fatalf("ERROR - DiscoverServiceDescription(): %v", err)
		}
	case len(*dev) > 0:
		desc, err = description.DiscoverDeviceDescription(*dev, *wait)
		if err != nil {
			log.Fatalf("ERROR - DiscoverDeviceDescription(): %v", err)
		}
	default:
		log.Fatal("Must provide -service or -device flag")
	}

	log.Printf("%+v", desc)
}
