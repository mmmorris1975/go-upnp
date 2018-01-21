package main

import (
	"flag"
	"github.com/mmmorris1975/upnp/description"
	"github.com/mmmorris1975/upnp/discovery"
	"log"
)

func main() {
	wait := flag.Duration("wait", discovery.DISCOVERY_WAIT_MAX_DURATION, "Duration for UPnP search request")
	svc := flag.String("service", "", "UPnP service to describe")
	dev := flag.String("device", "", "UPnP device to describe")
	flag.Parse()

	switch {
	case len(*svc) > 0:
		// Calls ServiceByType() for us
		svc, err := description.DiscoverServiceDescription(*svc, *wait)
		if err != nil {
			log.Fatalf("ERROR - DiscoverServiceDescription(): %v", err)
		}

		// Will be nil if nothing returned by search
		log.Printf("SERVICE: %+v", svc)
	case len(*dev) > 0:
		// Since a device can be searched by multiple dimensions, DiscoverDeviceDescription() returns
		// the top-level devices matched for the search, not the embedded device struct.  Calling
		// DeviceByType() can be used to navigate the returned device tree to extract the desired device
		// But doing so assumes that the argument passed to the -device flag is a device type URN and not
		// not something like a service URN.  However, if you know a device URN is passed, and the search
		// succeeds, you can be certain that the device you're looking for is contained somewhere in the
		// data returned by DiscoverDeviceDescription(), including any embedded sub-devices
		desc, err := description.DiscoverDeviceDescription(*dev, *wait)
		if err != nil {
			log.Fatalf("ERROR - DiscoverDeviceDescription(): %v", err)
		}
		if desc == nil {
			log.Fatal("ERROR - No results returned by device discovery")
		}

		d := desc.DeviceByType(*dev)
		if d == nil {
			d = desc.DeviceByService(*dev)
			if d == nil {
				log.Printf("WARNING - No device matching %s found, but search was successful, returning search result", *dev)
				log.Printf("DEVICE: %+v", desc)
				return
			}
		}
		log.Printf("DEVICE: %+v", d)
	default:
		log.Fatal("Must provide -service or -device flag")
	}
}
