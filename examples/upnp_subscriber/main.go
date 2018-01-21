package main

import (
	"flag"
	"github.com/mmmorris1975/upnp/description"
	"github.com/mmmorris1975/upnp/discovery"
	"github.com/mmmorris1975/upnp/eventing"
	"log"
)

func main() {
	wait := flag.Duration("wait", discovery.DISCOVERY_WAIT_MAX_DURATION, "Duration for UPnP search request")
	target := flag.String("target", discovery.DISCOVERY_TARGET_DEFAULT, "UPnP service to search for")
	flag.Parse()

	dd, err := description.DiscoverDeviceDescription(*target, *wait)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	s := dd.Device.ServiceByType(*target)

	u, err := dd.BuildURL(s.EventSubURL)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	log.Printf("EventSubUrl: %s", u)

	sm, err := eventing.NewSubscriptionManager(u, eventing.DEFAULT_SUBSCRIPTION_DURATION)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	ch := make(chan map[string]string, 10)
	go sm.EventLoop(ch)

	for e := range ch {
		log.Printf("EVENT: %s", e)
	}
}
