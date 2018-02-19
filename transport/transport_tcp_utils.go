package transport

import (
	"fmt"
	"log"
	"net"
	"strings"
)

// GetLocalAddresses returns the addresses TCP can listen to on the local machine
func GetLocalAddresses(port int) ([]string, error) {
	// go through all ifs
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	// gather addresses of all ifs
	ips := []net.Addr{}
	for _, iface := range ifaces {
		ifIPs, err := iface.Addrs()
		if err != nil {
			continue
		}
		ips = append(ips, ifIPs...)
	}

	// gather valid addresses
	addrs := []string{}
	for _, ip := range ips {
		cleanIP, valid := isValidIP(ip)
		if valid {
			hostPort := fmt.Sprintf("%s:%d", cleanIP, port)
			addrs = append(addrs, hostPort)
		}
	}
	return addrs, nil
}

// GetPublicAddresses returns the addresses TCP can listen to on the local machine
func GetPublicAddresses(port int, upnp UPNP) ([]string, error) {
	addrs := []string{}
	if upnp == nil {
		return addrs, nil
	}

	ip, err := upnp.ExternalIP()
	if err != nil {
		log.Println("External IP not found: ", err)
		return addrs, nil
	}

	if ip != "" {
		hostPort := fmt.Sprintf("%s:%d", ip, port)
		addrs = append(addrs, hostPort)
	}
	return addrs, nil
}

func isValidIP(addr net.Addr) (string, bool) {
	var ip net.IP
	switch v := addr.(type) {
	case *net.IPNet:
		ip = v.IP
	case *net.IPAddr:
		ip = v.IP
	}
	if ip == nil || ip.IsLoopback() || isIPv6(ip.String()) {
		return "", false
	}
	return ip.String(), true
}

func isIPv4(address string) bool {
	return strings.Count(address, ":") < 2
}

func isIPv6(address string) bool {
	return strings.Count(address, ":") >= 2
}
