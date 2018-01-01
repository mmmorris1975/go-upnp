package description

import (
	"encoding/xml"
	"net/url"
	"time"
)

// With the node added in getDescription(), this code should be UPnP 1.1 and 2.0 compliant

type Argument struct {
	XMLName              xml.Name `xml:"argument"`
	Name                 string   `xml:"name"`
	Direction            string   `xml:"direction"`
	RelatedStateVariable string   `xml:"relatedStateVariable"`
	RetVal               bool     `xml:"retval"`
}

type StateVariable struct {
	XMLName    xml.Name `xml:"stateVariable"`
	SendEvents string   `xml:"sendEvents,attr"`
	Multicast  string   `xml:"multicast,attr"`
	Name       string   `xml:"name"`
	DataType   string   `xml:"dataType"`
	// FIXME - can't use dataType>type,attr tag syntax
	//XmlType          string   `xml:"dataType>type,attr"`
	DefaultValue     string   `xml:"defaultValue"`
	MinValue         string   `xml:"allowedValueRange>minimum"`
	MaxValue         string   `xml:"allowedValueRange>maximum"`
	Step             string   `xml:"allowedValueRange>step"`
	AllowedValueList []string `xml:"allowedValueList>allowedValue"`
}

type Action struct {
	XMLName      xml.Name   `xml:"action"`
	Name         string     `xml:"name"`
	ArgumentList []Argument `xml:"argumentList>argument"`
}

// According to UPnP spec, section 2, services can supply additional attributes
// as part of ServiceDescription, but should be ignored when processing
type ServiceDescription struct {
	XMLName           xml.Name        `xml:"scpd"`
	ConfigId          int             `xml:"configId,attr"`
	UPnPMajorVersion  int             `xml:"specVersion>major"`
	UPnPMinorVersion  int             `xml:"specVersion>minor"`
	ActionList        []Action        `xml:"actionList>action"`
	ServiceStateTable []StateVariable `xml:"serviceStateTable>stateVariable"`
}

// Do a multicast discovery for the given service name and find the service description
// At this point, we only support getting the description for the 1st device returned from the search
func DiscoverServiceDescription(svcName string, wait time.Duration) (*ServiceDescription, error) {
	var svcUrl *url.URL

	dd, err := DiscoverDeviceDescription(svcName, wait)
	if err != nil {
		return nil, err
	}

	svc := dd.Device.ServiceByType(svcName)

	locUrl, err := url.Parse(device.Location)
	if err != nil {
		return nil, err
	}

	scpdUrl, err := url.Parse(svc.SCPDURL)
	if err != nil {
		return nil, err
	}

	// UPnP spec says that scpdUrl will always be relative to locUrl, so we
	// should never see a full http://server:port/path url in the SCPDURL field
	svcUrl = locUrl.ResolveReference(scpdUrl)
	return DescribeService(svcUrl.String())
}

// Perform service discovery for a given url (assumes discovery and device description already done)
func DescribeService(url string) (*ServiceDescription, error) {
	sd := &ServiceDescription{}

	err := getDescription(url, sd)
	if err != nil {
		return nil, err
	}

	return sd, nil
}
