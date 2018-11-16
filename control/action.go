package control

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"github.com/mmmorris1975/go-upnp/description"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"runtime"
)

type Action interface {
	// send action request, and provide return value in 'ret'
	Invoke(ret interface{}) error
}

type SimpleAction struct {
	XMLName xml.Name
	ctrlUrl *url.URL
	action  string
	service string
}

func NewAction(dd *description.DeviceDescription, svc, action string) (Action, error) {
	ctrl, err := getControlUrl(dd, svc)
	if err != nil {
		return nil, err
	}

	return newSimpleAction(ctrl, svc, action), nil
}

func newSimpleAction(ctrl *url.URL, svc, action string) *SimpleAction {
	a := new(SimpleAction)
	a.ctrlUrl = ctrl
	a.service = svc
	a.action = action
	a.XMLName = xml.Name{Space: svc, Local: action}
	return a
}

func (a *SimpleAction) Invoke(ret interface{}) error {
	req, err := a.buildSoapRequest()
	if err != nil {
		return err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	r := new(responseEnvelope)
	if err = xml.Unmarshal(b, r); err != nil {
		return err
	}

	if res.StatusCode == http.StatusOK {
		if ret != nil {
			if err := xml.Unmarshal(r.Body.Result, ret); err != nil {
				return err
			}
		}
	} else {
		f := new(Fault)
		if err := xml.Unmarshal(r.Body.Result, f); err != nil {
			return err
		}
		log.Printf("action returned HTTP %d: %s", res.StatusCode, f.Error())

		return f
	}

	return nil
}

func (a *SimpleAction) buildSoapRequest() (*http.Request, error) {
	e := NewEnvelope()
	e.Body.Action = a

	x, err := xml.Marshal(e)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, a.ctrlUrl.String(), bytes.NewBuffer(x))
	if err != nil {
		return nil, err
	}

	req.Header.Set("SOAPACTION", fmt.Sprintf(`"%s#%s"`, a.service, a.action))
	req.Header.Set("Content-Type", `text/xml; charset="utf-8"`)
	req.Header.Set("User-Agent", fmt.Sprintf(`%s/%s UPnP/1.1 xxx/1.0`, runtime.GOOS, runtime.Version()))

	return req, nil
}

func getControlUrl(dd *description.DeviceDescription, svc string) (*url.URL, error) {
	d := dd.DeviceByService(svc)
	if d == nil {
		return nil, fmt.Errorf("unable to find device for service: %s", svc)
	}

	var u string
	for _, e := range d.ServiceList {
		if svc == e.ServiceType {
			u = e.ControlURL
			break
		}
	}

	ctrl, err := dd.BuildURL(u)
	if err != nil {
		return nil, err
	}

	return ctrl, nil
}
