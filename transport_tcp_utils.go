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
			var ip net.IP
			switch v := iaddr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() || IsIPv6(ip.String()) {
				continue
			}
			addr := fmt.Sprintf("%s:%d", ip.String(), port)
			addrs = append(addrs, addr)
		}
	}

	return addrs, nil
}

func IsIPv4(address string) bool {
	return strings.Count(address, ":") < 2
}

func IsIPv6(address string) bool {
	return strings.Count(address, ":") >= 2
}
