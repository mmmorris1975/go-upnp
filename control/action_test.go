package control

import (
	"net/http"
	"net/url"
	"testing"
)

func TestAction(t *testing.T) {
	u, _ := url.Parse("http://localhost:12345/mock")
	a := newSimpleAction(u, "myService", "myAction")

	t.Run("request", func(t *testing.T) {
		r, err := a.buildSoapRequest()
		if err != nil {
			t.Fatal(err)
		}

		if r.Method != http.MethodPost {
			t.Error("request is not an HTTP POST")
		}

		if r.URL.String() != u.String() {
			t.Error("request URL mismatch")
		}

		if h := r.Header.Get("SOAPACTION"); len(h) < 1 {
			t.Error("SOAPACTION header missing, or not valid")
		}
		t.Log(r)
	})
}
