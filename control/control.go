package control

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"strings"
)

type Action interface{}

// While not technically an Action, we need the interface so we can be
// Unmarshaled as part of a SOAP response Body
type Fault struct {
	Action
	XMLName          xml.Name `xml:"s:Fault"`
	FaultCode        string   `xml:"faultcode"`
	FaultString      string   `xml:"faultstring"`
	ErrorCode        string   `xml:"detail>UPnPError>errorCode"`
	ErrorDescription string   `xml:"detail>UPnPError>errorDescription"`
}

type Body struct {
	XMLName xml.Name `xml:"s:Body"`
	Action
}

// Hack to deal with namespace prefixes
// https://github.com/golang/go/issues/9519
type Envelope struct {
	XMLName xml.Name `xml:"s:Envelope"`
	XMLNS   string   `xml:"xmlns:s,attr"`
	Body    Body
}

func parseXMLNameTag(action Action) (string, string) {
	f, ok := reflect.TypeOf(action).FieldByName("XMLName")
	if !ok {
		panic("missing XMLName field for Action struct")
	}

	t, ok := f.Tag.Lookup("xml")
	if !ok {
		panic("missing xml tag for XMLName field in Action struct")
	}

	parts := strings.SplitN(t, " ", 3)

	return parts[0], parts[1]
}

func NewEnvelope() Envelope {
	return Envelope{XMLNS: "http://schemas.xmlsoap.org/soap/envelope/"}
}

func Send(url string, action Action) (*http.Response, error) {
	e := NewEnvelope()
	e.Body.Action = action

	x, err := xml.Marshal(e)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(x))
	if err != nil {
		return nil, err
	}

	svc, act := parseXMLNameTag(action)
	req.Header.Set("SOAPACTION", fmt.Sprintf("\"%s#%s\"", svc, act))
	req.Header.Set("Content-Type", "text/xml; charset=\"utf-8\"")
	req.Header.Set("User-Agent", fmt.Sprintf("%s/%s UPnP/1.1 xxx/1.0", runtime.GOOS, runtime.Version()))

	fmt.Printf("%+v\n", req)

	res, err := http.DefaultClient.Do(req)
	/*
		if err != nil {
			// non-2xx HTTP status codes are not an error, this is for things like network errors
			return nil, err
		}

		if res.StatusCode < 400 {
			// request processed, or redirect
		} else {
			// request error (parse SOAP Fault)
		}
	*/

	return res, err
}
