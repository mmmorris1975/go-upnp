# UPnP Examples
`go test` is a great utility, and even has a facility to execute example code, however, in order to run the example tests
it expects very specific output, which is hard with a dynamic thing like UPnP device discovery on distinct networks (difficult
unless you write up a bunch of stubbing and mocking).  These examples are meant to be more instructive on how to write code
using this library, with the result being a set of (hopefully) useful utility programs that can be used outside of `go test`

## Tools
The following example programs are supplied.  (An example for performing UPnP control actions is not included as that requires
detailed knowledge of the specific service being called, which will vary between networks)

### upnp_discoverer
This tool will go out and discover UPnP devices on your local network via multicast discovery.  By default it will search for
targets using `ssdp:all`.  The search target and search duration can be controlled via command-line flags.  Also searching for
target `upnp:rootdevice` may yield "extra" devices which don't respond to `ssdp:all` search requests.

### upnp_describer
This tool will perform a UPnP multicast search for the specified device or service (provided via command-line flag), and obtain
the description data of that device or service, if found.

### upnp_subscriber
This tool will use the SubscriptionManager API to subscribe to events from the target specified by the UPnP search target flag.
It does not listen for multicast events, only for unicast events coming from the subscribed host.

### upnp_multicast_subscriber
This tool will listen for multicast UPnP event messages and print the event data.

## Building
The project top level make file can build these examples by running `make examples`, or you can use the Makefile inside the
examples directory by executing the default target via a simple call to `make`.  The resulting executables will be the name
of the tool with a `.cmd` extension (ex. upnp_discoverer.cmd)
