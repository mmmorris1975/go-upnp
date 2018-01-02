UPnP Golang Library
===================

A library in the Go language that support discovery, description, control, and eventing.
The library aims for UPnP 1.1 compliance, but should work with devices supporting 1.0
and 2.0 (without Device Protection)

Discovery
---------

The discovery module supports active unicast and multicast device discovery, as well as
passive multicast notification listening.  By default the `Discovery()` method of the
library will perform multicast service discovery to the `ssdp:all` target, which should
catch a vast majority of devices on the network.  NOTE, some devices don't respond to
the `ssdp:all` target, but do respond to a `upnp:rootdevice` discovery request.  To
customize the discovery request, modify the `discovery.SearchRequest` struct fields.

To passively listen for discovery notifications, run the `ListenNotify()` method and
discovered devices will be enumerated via the channel provided in the method call.

Description
-----------

The description module supports retrieval of device and service descriptions, converting
the returned SOAP XML into a struct.  For device description, call the `DescribeDevice()`
method with the full HTTP url provided by the Location field of a discovered device. A
convenience method called `DiscoverDeviceDescription()` is provided to perform discovery
for a given search target and extract the device description.  CAVEAT: only the
first device is used to perform the description, so you'll want to ensure the search
criteria only returns a single device.

For service description, call the `DescribeService()` method with the full HTTP url for
the SCPDURL provided by the device description.  This method assumes that device discovery
and device description has been performed prior to calling this method.  A convenience
method called `DiscoverServiceDescription()` is provided to perform the discovery and
device description in order to get the description for the service.  Provide the 
ServiceType as the parameter to the method to discover devices providing the service, and
extracting it's description.  The same caveat as `DiscoverDeviceDescription()` applies.

Control
-------

The control module provides a way to send UPnP control messages to devices without having
to worry (too much) about SOAP messaging.  By creating a struct which inherits the
`control.Action` interface, and marshals to valid XML needed for a UPnP action (see UPnP 1.1
spec, section 3.2.1), you can call the `Invoke()` method, specifying the device's full http 
ControlURL and the Action struct.  The value returned is a []byte of the response Body inner XML
so you are free to handle the data as you see fit, without all of the surrounding SOAP decoration.

Eventing
--------

The eventing module provides a way to subscribe and get notified for state change events from devices.
Unicast event subscription and multicast event listening is supported.

Unicast event subscription is handled through a SubscriptionManager, which handles the details of 
listening for event notifications as well as periodically refreshing the subscription, as required
by the UPnP spec.  Obtain a new SubscriptionManager instance by calling the `NewSubscriptionManager()` 
method, passing in the EventSubURL (obtained from the device description), and the desired subscription
lifetime.  Calling the `EventLoop()` method on the SubscriptionManager object will start the process
of listenting for events, returning state variables for the event as a map.  To cancel a subscription
call the `Unsubscribe()` method on the SubscriptionManager instance.

To receive events published via multicast, call the `ListenMulticastEvents()` method with a channel to
receive the events found. NOTE: this code has not been well tested

Building
--------

Contributing
------------

The usual github model for forking the repo and creating a pull request is the preferred way to
contribute to this tool.  Bug fixes, enhancements, doc updates, translations are always welcomed

References
----------
[UPnP 1.1 spec](http://upnp.org/specs/arch/UPnP-arch-DeviceArchitecture-v1.1.pdf)
