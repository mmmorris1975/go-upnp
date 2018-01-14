package discovery

import (
	"bufio"
	"bytes"
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

func search(req *SearchRequest) (*net.UDPAddr, error) {
	addr, err := getUDPAddr(req.Host, req.Port)
	if err != nil {
		return nil, err
	}

	// FIXME if host is multi-homed, we may want to explicitly set the local address
	c, err := net.DialUDP(addr.Network(), nil, addr)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	httpReq, err := http.NewRequest("M-SEARCH", "*", nil)
	if err != nil {
		return nil, err
	}
	httpReq.Host = addr.String()
	httpReq.Header.Set("MAN", "\"ssdp:discover\"")
	httpReq.Header.Set("ST", req.Target)
	httpReq.Header.Set("MX", strconv.Itoa(int(req.Wait.Seconds())))
	// UPnP 2.0 multicast search also MUST set CPFN.UPNP.ORG and MAY set CPUUID.UPNP.ORG to
	// set control point attributes used for Device Protection, not requred for unicast search

	buf := new(bytes.Buffer)
	httpReq.Write(buf)

	_, err = c.Write(buf.Bytes())
	if err != nil {
		return nil, err
	}

	l_addr, err := net.ResolveUDPAddr(c.LocalAddr().Network(), c.LocalAddr().String())
	if err != nil {
		return nil, err
	}

	return l_addr, nil
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

func readHttpRequest(c *net.UDPConn) (*http.Request, error) {
	rdr := bufio.NewReader(c)

	r, err := http.ReadRequest(rdr)
	if err != nil {
		err = handleHttpError(err)
		return nil, err
	}

	return r, nil
}

func readHttpResponse(c *net.UDPConn) (*http.Response, error) {
	rdr := bufio.NewReader(c)

	r, err := http.ReadResponse(rdr, nil)
	if err != nil {
		err = handleHttpError(err)
		return nil, err
	}

	return r, nil
}

func getSearchResponses(addr *net.UDPAddr, wait time.Duration, ch chan<- *SearchResponse) error {
	c, err := net.ListenUDP(addr.Network(), addr)
	if err != nil {
		log.Printf("ERROR - ListenUDP(): %v\n", err)
		close(ch)
		return err
	}
	defer c.Close()

	err = c.SetReadDeadline(time.Now().Add(wait).Add(1 * time.Second))
	if err != nil {
		log.Printf("ERROR - SetReadDeadline(): %v\n", err)
		close(ch)
		return err
	}

	for true {
		r, err := readHttpResponse(c)
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

		ch <- parseSearchResponse(r)
	}

	close(ch)
	return nil
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

	addr, err := search(req)
	if err != nil {
		log.Printf("ERROR - Discover(): %s", err)
		close(ch)
		return err
	}

	go getSearchResponses(addr, req.Wait, ch)
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

	for true {
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
