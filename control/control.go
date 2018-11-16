package control

import (
	"encoding/xml"
	"fmt"
)

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
