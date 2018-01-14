package eventing

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
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
	Result  Result   `xml:",any"`
}

type Result struct {
	XMLName xml.Name
	Value   string `xml:",chardata"`
}

// structs and helper to send a multicast event xml
type eventData struct {
	XMLName    xml.Name `xml:"e:propertyset"`
	XMLNS      string   `xml:"xmlns:e,attr"`
	Properties []property
}

type property struct {
	XMLName xml.Name `xml:"e:property"`
	Result  Result
}

func NewEventData() *eventData {
	return &eventData{XMLNS: "urn:schemas-upnp-org:event-1-0"}
}

func ListenMulticastEvents(ch chan<- *Event) error {
	addr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(MCAST_EVENT_ADDR, strconv.Itoa(MCAST_EVENT_PORT)))
	if err != nil {
		log.Printf("ERROR - ResolveUDPAddr(): %s", err)
		close(ch)
		return err
	}

	c, err := net.ListenMulticastUDP(addr.Network(), nil, addr)
	if err != nil {
		log.Printf("ERROR - ListenMulticastUDP(): %s", err)
		close(ch)
		return err
	}
	defer c.Close()

	for true {
		e, err := readEvent(c)
		if err != nil {
			log.Printf("ERROR - readEvent(): %v\n", err)
			close(ch)
			return err
		}

		ch <- e
	}

	close(ch)
	return nil
}

func SendMulticastEvent(h *EventHeader, r *[]Result, laddr *net.UDPAddr) error {
	addr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(MCAST_EVENT_ADDR, strconv.Itoa(MCAST_EVENT_PORT)))
	if err != nil {
		log.Printf("ERROR - ResolveUDPAddr(): %s", err)
		return err
	}

	if len(h.LVL) < 1 {
		h.LVL = "upnp:/info"
	}

	e := NewEventData()
	for _, j := range *r {
		e.Properties = append(e.Properties, property{Result: j})
	}

	b, err := xml.Marshal(e)
	if err != nil {
		log.Printf("ERROR - Marshal(): %v", err)
		return err
	}

	req, err := http.NewRequest("NOTIFY", "*", bytes.NewBuffer(b))
	if err != nil {
		log.Printf("ERROR - NewRequest(): %s", err)
		return err
	}
	req.Host = addr.String()
	req.Header.Set("Content-Type", "text/xml; charset=\"utf-8\"")
	req.Header.Set("NT", "upnp:event")
	req.Header.Set("NTS", "upnp:propchange")
	req.Header.Set("LVL", h.LVL)
	req.Header.Set("USN", h.USN)
	req.Header.Set("SVCID", h.SVCID)
	req.Header.Set("SEQ", strconv.Itoa(h.SEQ))
	req.Header.Set("BOOTID.UPNP.ORG", strconv.Itoa(h.BootId))

	// TESTING
	log.Printf("REQ: %+v", req)
	log.Printf("EVT: %s", string(b))
	return nil

	// TODO - send stuff
	c, err := net.DialUDP(addr.Network(), laddr, addr)
	if err != nil {
		log.Printf("ERROR - DialUDP(): %v", err)
		return err
	}
	defer c.Close()

	if err := req.Write(c); err != nil {
		log.Printf("ERROR - Write(): %v", err)
		return err
	}

	return nil
}

func readEvent(c *net.UDPConn) (*Event, error) {
	rdr := bufio.NewReader(c)

	r, err := http.ReadRequest(rdr)
	if err != nil {
		err = handleHttpError(err)
		return nil, err
	}
	if r == nil {
		return nil, nil
	}
	if r.Body != nil {
		defer r.Body.Close()
	}

	h := EventHeader{}
	h.NT = r.Header.Get("NT")
	h.NTS = r.Header.Get("NTS")
	h.USN = r.Header.Get("USN")
	h.SVCID = r.Header.Get("SVCID")
	h.LVL = r.Header.Get("LVL")

	seq, err := strconv.Atoi(r.Header.Get("SEQ"))
	if err != nil {
		seq = 0
	}
	h.SEQ = seq

	bid, err := strconv.Atoi(r.Header.Get("BOOTID.UPNP.ORG"))
	if err != nil {
		bid = 0
	}
	h.BootId = bid

	d := EventData{}
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	err = xml.Unmarshal(b, &d)
	if err != nil {
		return nil, err
	}

	return &Event{EventHeader: h, EventData: d}, nil
}

func handleHttpError(err error) error {
	switch t := err.(type) {
	case *net.OpError:
		if !t.Temporary() {
			return err
		}
	default:
		return err
	}

	return nil
}
