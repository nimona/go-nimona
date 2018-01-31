package fabric

import (
	"fmt"
	"net"
	"strings"
)

// GetAddresses returns the addresses TCP can listen to on the local machine
func GetAddresses(port int) ([]string, error) {
	// add all addresses to peer
	addrs := []string{}

	// go through all ifs
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	// and find their addresses
	for _, i := range ifaces {
		iaddrs, err := i.Addrs()
		if err != nil {
			continue
		}
		for _, iaddr := range iaddrs {
			if isValidIP(iaddr) {
				addrs = append(addrs, fmt.Sprintf("%s:%d", iaddr.String(), port))
			}
		}
	}

	return addrs, nil
}

func isValidIP(addr net.Addr) bool {
	var ip net.IP
	switch v := addr.(type) {
	case *net.IPNet:
		ip = v.IP
	case *net.IPAddr:
		ip = v.IP
	}
	if ip == nil || ip.IsLoopback() || isIPv6(ip.String()) {
		return false
	}
	return true
}

func isIPv4(address string) bool {
	return strings.Count(address, ":") < 2
}

func isIPv6(address string) bool {
	return strings.Count(address, ":") >= 2
}
