package main

import (
	"flag"
	"github.com/mmmorris1975/go-upnp/description"
	"github.com/mmmorris1975/go-upnp/discovery"
	"github.com/mmmorris1975/go-upnp/eventing"
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

	s := dd.ServiceByType(*target)
	if s == nil {
		log.Fatalf("No service named %s found in device %s", *target, dd.Device.DeviceType)
	}

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
