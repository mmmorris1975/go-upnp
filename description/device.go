package description

import (
	"encoding/xml"
	"github.com/mmmorris1975/upnp/discovery"
	"net/url"
	"time"
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
	DeviceList       []Device  `xml:"deviceList>device"`
	PresentationURL  string    `xml:"presentationURL"`
}

var iconCache map[string]Icon

func (d *Device) IconByMimetype(mt string) Icon {
	if len(iconCache) < 1 {
		iconCache = make(map[string]Icon, len(d.IconList))
		for _, e := range d.IconList {
			iconCache[e.Mimetype] = e
		}
	}

	return iconCache[mt]
}

func (d *Device) ServiceByType(st string) *Service {
	var svc *Service

	for _, s := range d.ServiceList {
		if string(st) == string(s.ServiceType) {
			svc = &s
			break
		}
	}

	for _, e := range d.DeviceList {
		s := e.ServiceByType(st)
		if s != nil {
			svc = s
			break
		}
	}

	return svc
}

func (d *Device) DeviceByType(dt string) *Device {
	if dt == d.DeviceType {
		return d
	}

	for _, e := range d.DeviceList {
		x := e.DeviceByType(dt)
		if x != nil {
			return x
		}
	}

	return nil
}

// According to UPnP spec, section 2, devices can supply additional attributes
// as part of Device or DeviceDescription, but should be ignored when processing
type DeviceDescription struct {
	XMLName          xml.Name `xml:"root"`
	ConfigId         int      `xml:"configId,attr"`
	UPnPMajorVersion int      `xml:"specVersion>major"`
	UPnPMinorVersion int      `xml:"specVersion>minor"`
	URLBase	         string   `xml:"URLBase"`  // UPnP 1.0, deprecated in UPnp 1.1
	Device           Device
	location         *url.URL
}

func (d *DeviceDescription) DeviceByType(dt string) *Device {
	return d.Device.DeviceByType(dt)
}

func (d *DeviceDescription) ServiceByType(st string) *Service {
	return d.Device.ServiceByType(st)
}

func (d *DeviceDescription) BuildURL(path string) (*url.URL, error) {
	p, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	u := d.location.ResolveReference(p)
	return u, nil
}

// Do a multicast discovery for the given ssdp target and find the device description
// At this point, we only support getting the description for the 1st device returned from the search
func DiscoverDeviceDescription(target string, wait time.Duration) (*DeviceDescription, error) {
	ch := make(chan *discovery.SearchResponse, 10)
	doDiscovery(target, wait, ch)

	device := <-ch
	if device == nil {
		return nil, nil
	}

	dd, err := DescribeDevice(device.Location)
	if err != nil {
		return nil, err
	}

	return dd, nil
}

// Describe the top-level device at the given URL, although this will probably always be the
// Location field of a UPnP discovery (since devices themselves don't contain description URLs)
func DescribeDevice(u string) (*DeviceDescription, error) {
	dd := &DeviceDescription{}

	err := getDescription(u, dd)
	if err != nil {
		return nil, err
	}

	// Honor UPnP 1.0 URLBase attribute if present, otherwise use the URL we were passed
	if len(dd.URLBase) > 0 {
		u = dd.URLBase
	}

	o, err := url.Parse(u)
	if err != nil {
		return nil, err
	}
	dd.location = o

	return dd, nil
}
