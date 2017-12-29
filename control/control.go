package control

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"runtime"
	"strings"
)

type Action interface{}

type Fault struct {
	XMLName          xml.Name `xml:"Fault"`
	FaultCode        string   `xml:"faultcode"`
	FaultString      string   `xml:"faultstring"`
	ErrorCode        string   `xml:"detail>UPnPError>errorCode"`
	ErrorDescription string   `xml:"detail>UPnPError>errorDescription"`
}

func (f Fault) Error() string {
	return fmt.Sprintf("%v: %v", f.ErrorDescription, f.ErrorCode)
}

type Body struct {
	XMLName xml.Name `xml:"s:Body"`
	Action
}

// Hack to deal with XML namespace prefixes
// https://github.com/golang/go/issues/9519
type Envelope struct {
	XMLName xml.Name `xml:"s:Envelope"`
	XMLNS   string   `xml:"xmlns:s,attr"`
	Body    Body
}

// Raw HTML/XML response will have namespace-prefixed elements but Go XML
// processing ignores those, so we need distinct structs for request & response
type responseEnvelope struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    responseBody
}

type responseBody struct {
	XMLName xml.Name `xml:"Body"`
	Result  []byte   `xml:",innerxml"`
}

func NewEnvelope() Envelope {
	return Envelope{XMLNS: "http://schemas.xmlsoap.org/soap/envelope/"}
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

func buildSoapRequest(url string, action Action) (*http.Request, error) {
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

	return req, nil
}

// If request was successful, return a []byte of the innerxml of the Body response element
// This allows the caller to deal with the returned information, stripped of the surrounding
// SOAP Envelope and Body decorations.
func Invoke(url string, action Action) ([]byte, error) {
	req, err := buildSoapRequest(url, action)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		// non-2xx HTTP status codes are not an error, this is for things like network errors
		return nil, err
	}
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	r := responseEnvelope{}
	err = xml.Unmarshal(b, &r)
	if err != nil {
		return nil, err
	}

	if res.StatusCode < 400 {
		// request processed, or redirect
		return r.Body.Result, nil
	} else {
		// request error (parse SOAP Fault)
		f := Fault{}
		err = xml.Unmarshal(r.Body.Result, &f)
		if err != nil {
			return nil, err
		}

		return nil, f
	}
}
