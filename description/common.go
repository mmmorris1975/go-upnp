package description

import (
	"bytes"
	"encoding/xml"
	"github.com/mmmorris1975/upnp/discovery"
	"net/http"
	"time"
)

func getDescription(url string, v interface{}) error {
	// UPnP 2.0 HTTP requests also MUST set CPFN.UPNP.ORG and MAY set CPUUID.UPNP.ORG
	// headers to set control point attributes used for Device Protection
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	if res.Body != nil {
		defer res.Body.Close()
	}

	buf := bytes.NewBuffer(make([]byte, 0, 10240))
	_, err = buf.ReadFrom(res.Body)
	if err != nil {
		return err
	}

	err = xml.Unmarshal(buf.Bytes(), v)
	if err != nil {
		return err
	}

	return nil
}

func doDiscovery(target string, wait time.Duration, ch chan<- *discovery.SearchResponse) {
	// multicast discovery
	discReq := discovery.NewSearchRequest()
	discReq.Target = target
	discReq.Wait = wait

	discovery.Discover(discReq, ch)
}
