package discovery

import (
	"log"
	"testing"
	"time"
)

func TestDiscover(t *testing.T) {
	t.Run("multicast", func(t *testing.T) {
		ch := make(chan *SearchResponse, 10)
		s := NewSearchRequest()

		Discover(s, ch)
		for {
			v, ok := <-ch
			if !ok {
				break
			}
			log.Printf("%v", v)
		}
	})

	t.Run("unicast", func(t *testing.T) {
		ch := make(chan *SearchResponse, 10)
		s := NewSearchRequest()
		s.Host = "192.168.1.1" // FIXME - non-constant test param
		s.Target = "upnp:rootdevice"
		s.Wait = 2 * time.Second

		Discover(s, ch)
		for {
			v, ok := <-ch
			if !ok {
				break
			}
			log.Printf("%v", v)
		}
	})

	t.Run("notify", func(t *testing.T) {
		// FIXME - how can we make this time out so we're not waiting forever for 3 responses?
		ch := make(chan *NotifyResponse, 10)
		ListenNotify(ch)

		for i := 0; i < 3; i++ {
			log.Printf("%v", <-ch)
		}
	})
}
