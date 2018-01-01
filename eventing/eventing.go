package eventing

import (
	"encoding/xml"
)

const (
	MCAST_EVENT_ADDR = "239.255.255.246"
	MCAST_EVENT_PORT = 7900
)

type EventHeader struct {
	NT  string
	NTS string
	SID string
	SEQ int

	// multicast event headers
	USN    string
	SVCID  string
	LVL    string
	BootId int
}

type EventData struct {
	XMLName    xml.Name   `xml:"propertyset"`
	Properties []Property `xml:"property"`
}

type Event struct {
	EventHeader
	EventData
}

type Property struct {
	XMLName xml.Name `xml:"property"`
	Result  string   `xml:",innerxml"`
}

func ListenMulticastEvents() {
	/*
	 */
}
