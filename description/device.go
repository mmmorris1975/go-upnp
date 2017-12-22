package description

import (
	"encoding/xml"
)

// With the node added in getDescription(), this code should be UPnP 1.1 and 2.0 compliant

type Icon struct {
	XMLName  xml.Name `xml:"icon"`
	Mimetype string   `xml:"mimetype"`
	Width    int      `xml:"width"`
	Height   int      `xml:"height"`
	Depth    int      `xml:"depth"`
	URL      string   `xml:"url"`
}

type Service struct {
	XMLName     xml.Name `xml:"service"`
	ServiceType string   `xml:"serviceType"`
	ServiceId   string   `xml:"serviceId"`
	SCPDURL     string   `xml:"SCPDURL"`
	ControlURL  string   `xml:"controlURL"`
	EventSubURL string   `xml:"eventSubURL"`
}

type Device struct {
	XMLName          xml.Name  `xml:"device"`
	DeviceType       string    `xml:"deviceType"`
	FriendlyName     string    `xml:"friendlyName"`
	Manufacturer     string    `xml:"manufacturer"`
	ManufacturerURL  string    `xml:"manufacturerURL"`
	ModelDescription string    `xml:"modelDescription"`
	ModelName        string    `xml:"modelName"`
	ModelNumber      string    `xml:"modelNumber"`
	ModelURL         string    `xml:"modelURL"`
	SerialNumber     string    `xml:"serialNumber"`
	UDN              string    `xml:"UDN"`
	UPC              string    `xml:"UPC"`
	IconList         []Icon    `xml:"iconList>icon"`
	ServiceList      []Service `xml:"serviceList>service"`
	DeviceList       []Device  `xml:"deviceList"`
	PresentationURL  string    `xml:"presentationURL"`
}

// According to UPnP spec, section 2, devices can supply additional attributes
// as part of Device or DeviceDescription, but should be ignored when processing
type DeviceDescription struct {
	XMLName          xml.Name `xml:"root"`
	ConfigId         int      `xml:"configId,attr"`
	UPnPMajorVersion int      `xml:"specVersion>major"`
	UPnPMinorVersion int      `xml:"specVersion>minor"`
	Device           Device
}

func GetDeviceDescription(url string) (*DeviceDescription, error) {
	dd := &DeviceDescription{}

	err := getDescription(url, dd)
	if err != nil {
		return nil, err
	}

	return dd, nil
}
