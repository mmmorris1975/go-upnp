package discovery

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"
)

// With the note added below in search(), this code should be UPnP 1.1 and 2.0 compliant

const (
	// Interesting note, some devices like WeMo switches don't respond to ssdp:all queries, but will
	// answer upnp:rootdevice queries.  Something to keep in mind, I'm sure it's not the only case
	DISCOVERY_ADDR_DEFAULT      = "239.255.255.250"
	DISCOVERY_PORT_DEFAULT      = 1900
	DISCOVERY_TARGET_DEFAULT    = "ssdp:all"
	DISCOVERY_WAIT_MIN_DURATION = 1 * time.Second
	DISCOVERY_WAIT_MAX_DURATION = 5 * time.Second
)

var Logger *log.Logger

type Discoverer interface {
	Discover(req *SearchRequest, ch chan<- *SearchResponse)
	ListenNotify(ch chan<- *NotifyResponse)
}

type SSDPResponse struct {
	Location     string
	CacheControl string
	Server       string
	USN          string
	BootId       int
	ConfigId     int
	SearchPort   int
}

type NotifyResponse struct {
	SSDPResponse
	NT         string
	NTS        string
	NextBootId int
}

type SearchResponse struct {
	SSDPResponse
	ST string
}

type SearchRequest struct {
	Host   string
	Port   int
	Target string
	Wait   time.Duration
}

func NewSearchRequest() *SearchRequest {
	return &SearchRequest{
		Host:   DISCOVERY_ADDR_DEFAULT,
		Port:   DISCOVERY_PORT_DEFAULT,
		Target: DISCOVERY_TARGET_DEFAULT,
		Wait:   DISCOVERY_WAIT_MAX_DURATION,
	}
}

func getUDPAddr(host string, port int) (*net.UDPAddr, error) {
	addr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(host, strconv.Itoa(port)))
	if err != nil {
		return nil, err
	}

	return addr, nil
}

func parseSSDPResponse(h *http.Header) SSDPResponse {
	r := SSDPResponse{}

	r.Location = h.Get("Location")
	r.CacheControl = h.Get("Cache-Control")
	r.Server = h.Get("Server")
	r.USN = h.Get("USN")

	bootId, err := strconv.Atoi(h.Get("BOOTID.UPNP.ORG"))
	if err == nil {
		r.BootId = bootId
	}

	configId, err := strconv.Atoi(h.Get("CONFIGID.UPNP.ORG"))
	if err == nil {
		r.ConfigId = configId
	}

	searchPort, err := strconv.Atoi(h.Get("SEARCHPORT.UPNP.ORG"))
	if err == nil {
		r.SearchPort = searchPort
	}

	return r
}

func parseSearchResponse(r *http.Response) *SearchResponse {
	sr := &SearchResponse{}
	sr.SSDPResponse = parseSSDPResponse(&r.Header)

	sr.ST = r.Header.Get("ST")

	return sr
}

func parseNotifyResponse(r *http.Request) *NotifyResponse {
	nr := &NotifyResponse{}
	nr.SSDPResponse = parseSSDPResponse(&r.Header)

	nr.NT = r.Header.Get("NT")
	nr.NTS = r.Header.Get("NTS")

	nextBootId, err := strconv.Atoi(r.Header.Get("NEXTBOOTID.UPNP.ORG"))
	if err == nil {
		nr.NextBootId = nextBootId
	}
	// UPnP 2.0 ssdp:alive message may contain SECURELOCATION.UPNP.ORG header
	// if device protection is being used

	return nr
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

func readHttpRequest(rdr io.Reader) (*http.Request, error) {
	r, err := http.ReadRequest(bufio.NewReader(rdr))
	if err != nil {
		err = handleHttpError(err)
		return nil, err
	}

	return r, nil
}

func readHttpResponse(rdr io.Reader) (*http.Response, error) {
	r, err := http.ReadResponse(bufio.NewReader(rdr), nil)
	if err != nil {
		err = handleHttpError(err)
		return nil, err
	}

	return r, nil
}

func search(req *SearchRequest, ch chan<- *SearchResponse) {
	a, err := getUDPAddr(req.Host, req.Port)
	if err != nil {
		log.Printf("ERROR - getUDPAddr(): %s", err)
		close(ch)
		return
	}

	c, err := net.ListenPacket(a.Network(), ":0")
	if err != nil {
		log.Printf("ERROR - ListenPacket(): %v", err)
		close(ch)
		return
	}

	httpReq, err := http.NewRequest("M-SEARCH", "*", nil)
	if err != nil {
		log.Printf("ERROR - http NewRequest(): %v", err)
		close(ch)
		return
	}
	httpReq.Host = a.String()
	httpReq.Header.Set("MAN", "\"ssdp:discover\"")
	httpReq.Header.Set("ST", req.Target)
	httpReq.Header.Set("MX", strconv.Itoa(int(req.Wait.Seconds())))
	// UPnP 2.0 multicast search also MUST set CPFN.UPNP.ORG and MAY set CPUUID.UPNP.ORG to
	// set control point attributes used for Device Protection, not required for unicast search

	buf := new(bytes.Buffer)
	httpReq.Write(buf)

	c.SetReadDeadline(time.Now().Add(req.Wait))
	go getSearchResponses(c, ch)

	if _, err := c.WriteTo(buf.Bytes(), a); err != nil {
		log.Printf("ERROR - WriteTo(): %v", err)
		close(ch)
		return
	}
}

func getSearchResponses(c net.PacketConn, ch chan<- *SearchResponse) {
	defer close(ch)
	defer c.Close()

	for {
		b := make([]byte, 4096)

		_, _, err := c.ReadFrom(b)
		if err != nil {
			switch t := err.(type) {
			case *net.OpError:
				if t.Timeout() {
					break
				}
			default:
				log.Printf("ERROR - ReadFrom(): %v", err)
			}
			return
		}

		r, err := readHttpResponse(bytes.NewReader(b))
		if err != nil {
			log.Printf("ERROR - ReadHttpResponse(): %v", err)
			return
		}
		if r == nil {
			break
		}
		if r.Body != nil {
			defer r.Body.Close()
		}

		ch <- parseSearchResponse(r)
	}
}

func doLog(fmt string, vars ...interface{}) {
	if Logger != nil {
		Logger.Printf(fmt, vars...)
	}
}

func Discover(req *SearchRequest, ch chan<- *SearchResponse) error {
	waitSec := req.Wait.Seconds()
	if waitSec < DISCOVERY_WAIT_MIN_DURATION.Seconds() {
		log.Printf("WARNING - Provided wait time of %0.3f seconds is less than allowed value of 1s, raising to 1s\n", waitSec)
		req.Wait = 1 * time.Second
	}

	if waitSec > DISCOVERY_WAIT_MAX_DURATION.Seconds() {
		log.Printf("WARNING - Provided wait time of %0.3f seconds is more than allowed value of 5s, lowering to 5s\n", waitSec)
		req.Wait = 5 * time.Second
	}

	go search(req, ch)
	return nil
}

func ListenNotify(ch chan<- *NotifyResponse) error {
	addr, err := getUDPAddr(DISCOVERY_ADDR_DEFAULT, DISCOVERY_PORT_DEFAULT)
	if err != nil {
		log.Printf("ERROR - getUDPAddr(): %s", err)
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

	for {
		r, err := readHttpRequest(c)
		if err != nil {
			log.Printf("ERROR - readHttpResponse(): %v\n", err)
			close(ch)
			return err
		}
		if r == nil {
			break
		}
		if r.Body != nil {
			defer r.Body.Close()
		}

		ch <- parseNotifyResponse(r)
	}

	close(ch)
	return nil
}
