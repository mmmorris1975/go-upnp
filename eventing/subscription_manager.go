package eventing

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	DEFAULT_SUBSCRIPTION_DURATION = 30 * time.Minute
	MIN_SUBSCRIPTION_DURATION     = 1 * time.Second
	// Devices will provide a default value if subscription duration
	// exceeds their internal max allowed value
	// practical limit on duration value is (2^31 - 1)
)

type SubscriptionManager struct {
	URL      *url.URL
	SID      string
	Lifetime time.Duration
	listener net.Listener
}

func (m *SubscriptionManager) EventLoop(ch chan<- map[string]string) {
	// start NOTIFY handler to recieve event data in background,
	// needs to be listening before sending subscription request
	events := make(chan *Event, 10)
	go notifyHandler(m.listener, events)

	// start goroutine to do initial subscription request, and
	// manage resubscription activities
	go m.manageSubscription()

	for e := range events {
		m := make(map[string]string)
		for _, p := range e.Properties {
			m[p.Result.XMLName.Local] = p.Result.Value
		}
		ch <- m
	}
}

func (m *SubscriptionManager) Unsubscribe() error {
	req, err := http.NewRequest("UNSUBSCRIBE", m.URL.String(), http.NoBody)
	if err != nil {
		return err
	}

	if len(m.SID) > 0 {
		req.Header.Set("SID", m.SID)
	} else {
		// no SID, just return
		return nil
	}

	c := http.Client{}
	res, err := c.Do(req)
	if err != nil {
		// not http errors
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("UNSUBSCRIBE request returned HTTP %d", res.StatusCode)
	}

	return nil
}

func (m *SubscriptionManager) manageSubscription() error {
	req, err := m.newSubscriptionRequest()
	if err != nil {
		log.Printf("ERROR - Unable to build subscription request: %v", err)
		return err
	}

	if len(m.SID) < 1 {
		req.Header.Set("NT", "upnp:event")
		req.Header.Set("CALLBACK", fmt.Sprintf("<http://%s/>", m.listener.Addr()))
	} else {
		req.Header.Set("SID", m.SID)

		renewTime := m.Lifetime.Seconds() * 0.9
		time.Sleep(time.Duration(renewTime) * time.Second)
		log.Printf("INFO - Renewing subscription for SID %s", m.SID)
	}

	err = m.doSubscriptionRequest(req)
	if err != nil {
		log.Printf("ERROR - subscription error: %v", err)
		return err
	}

	return m.manageSubscription()
}

func (m *SubscriptionManager) newSubscriptionRequest() (*http.Request, error) {
	req, err := http.NewRequest("SUBSCRIBE", m.URL.String(), http.NoBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("TIMEOUT", fmt.Sprintf("Second-%d", int(m.Lifetime.Seconds())))

	return req, nil
}

func (m *SubscriptionManager) doSubscriptionRequest(req *http.Request) error {
	c := http.Client{}
	res, err := c.Do(req)
	if err != nil {
		// not http errors
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("SUBSCRIBE request returned HTTP %d", res.StatusCode)
	}

	tmout := -1
	h := res.Header.Get("TIMEOUT")
	if len(h) > 0 {
		t := strings.Split(h, "-")
		tmout, _ = strconv.Atoi(t[1])
	}

	m.SID = res.Header.Get("SID")
	m.Lifetime = time.Duration(tmout) * time.Second

	return nil
}

// FIXME - url may become invalid if device we're subscribing to restarts and the subscription url changes.
// Change this to the service name we want to subscribe to, and deal with the discovery and url building internally
// otherwise our subscription loops may crash when it attempts to resubscribe to a dead endpoint
// Provide ability to filter discovery/description data, in case more than 1 is found?
// - description.DiscoverDeviceDescription() only returns the 1st discovered device, so either we modify that
//   behavior, or we just go with it, and assume our service name is targeted enough to find the right one
func NewSubscriptionManager(url *url.URL, exp time.Duration) (*SubscriptionManager, error) {
	if exp < MIN_SUBSCRIPTION_DURATION {
		log.Printf("WARNING - provided subscription duration less than allowed minimum duration (%s), using default of %s",
			MIN_SUBSCRIPTION_DURATION.String(), DEFAULT_SUBSCRIPTION_DURATION.String())
		exp = DEFAULT_SUBSCRIPTION_DURATION
	}

	l, err := setupListener(url)
	if err != nil {
		return nil, err
	}

	s := SubscriptionManager{
		URL:      url,
		Lifetime: exp,
		listener: l,
	}

	return &s, nil
}

func setupListener(url *url.URL) (net.Listener, error) {
	// Determine which local address will be used for the subscription requests
	c, err := net.Dial("tcp", url.Host)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	h, _, err := net.SplitHostPort(c.LocalAddr().String())
	if err != nil {
		return nil, err
	}

	// Instantiate listener on ephemeral port
	l, err := net.Listen("tcp", fmt.Sprintf("%s:0", h))
	if err != nil {
		return nil, err
	}

	return l, nil
}

func notifyHandler(l net.Listener, ch chan<- *Event) error {
	// Always respond with HTTP 200 (even if error, since it's likely a problem on our end)
	// Simply log any errors and bail out of the request, so we still get future notifications
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		seq, err := strconv.Atoi(r.Header.Get("SEQ"))
		if err != nil {
			seq = 0
		}

		h := EventHeader{
			NT:  r.Header.Get("NT"),
			NTS: r.Header.Get("NTS"),
			SID: r.Header.Get("SID"),
			SEQ: seq,
		}

		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Printf("ERROR - Failed to read notification body: %v", err)
			return
		}

		d := EventData{}
		if err = xml.Unmarshal(b, &d); err != nil {
			log.Printf("ERROR - Unmarshal(): %v", err)
			return
		}

		e := Event{EventHeader: h, EventData: d}
		ch <- &e
	})

	return http.Serve(l, nil)
}
