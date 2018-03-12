package net

// UPNP interface
type UPNP interface {
	ExternalIP() (string, error)
	Forward(port uint16, desc string) error
	Clear(port uint16) error
	Location() string
}
