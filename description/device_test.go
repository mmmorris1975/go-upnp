package description

import (
	"github.com/mmmorris1975/go-upnp/discovery"
	"testing"
	"time"
)

func TestDiscoverDeviceDescription(t *testing.T) {
	t.Run("upnp:rootdevice", func(t *testing.T) {
		dd, err := DiscoverDeviceDescription("upnp:rootdevice", 5*time.Second)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(dd)
	})
}

func TestDescribeDevice(t *testing.T) {
	t.Run("1", func(t *testing.T) {
		ch := make(chan *discovery.SearchResponse, 1)
		s := discovery.NewSearchRequest()
		s.Host = "192.168.1.1" // FIXME - non-constant test param
		s.Target = "urn:schemas-upnp-org:service:WANCommonInterfaceConfig:1"
		s.Wait = 2 * time.Second
		discovery.Discover(s, ch)

		r := <-ch
		if r == nil {
			t.Fatal("No device found")
		}

		dd, err := DescribeDevice(r.Location)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(dd)
	})
}
