package net

import (
	"fmt"
	"net"
	"strings"
)

// GetAddresses returns the addresses the transport is listening to
func GetAddresses(
	protocol string,
	l net.Listener,
	includeLocal bool,
	includePrivate bool,
	includeIPV6 bool,
) []string {
	port := l.Addr().(*net.TCPAddr).Port
	// TODO log errors
	addrs, _ := GetLocalPeerAddresses(
		port,
		includeLocal,
		includePrivate,
		includeIPV6,
	)
	for i, addr := range addrs {
		addrs[i] = fmt.Sprintf("%s:%s", protocol, addr)
	}
	return addrs
}

// GetLocalPeerAddresses returns the addresses TCP can listen to on the
// local machine
func GetLocalPeerAddresses(
	port int,
	includeLocal bool,
	includePrivate bool,
	includeIPV6 bool,
) ([]string, error) {
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
		cleanIP, valid := isValidIP(
			ip,
			includeLocal,
			includePrivate,
			includeIPV6,
		)
		if valid {
			hostPort := fmt.Sprintf("%s:%d", cleanIP, port)
			addrs = append(addrs, hostPort)
		}
	}
	return addrs, nil
}

func isValidIP(
	addr net.Addr,
	includeLocal bool,
	includePrivate bool,
	includeIPV6 bool,
) (string, bool) {
	var ip net.IP
	switch v := addr.(type) {
	case *net.IPNet:
		ip = v.IP
	case *net.IPAddr:
		ip = v.IP
	}
	if ip == nil {
		return "", false
	}
	if !includeLocal && ip.IsLoopback() {
		return "", false
	}
	if !includePrivate && isPrivate(ip) {
		return "", false
	}
	if !includeIPV6 && isIPv6(ip.String()) {
		return "", false
	}
	return ip.String(), true
}

func isIPv6(address string) bool {
	return strings.Count(address, ":") >= 2
}

func isPrivate(ip net.IP) bool {
	_, block24, _ := net.ParseCIDR("10.0.0.0/8")
	_, block20, _ := net.ParseCIDR("172.16.0.0/12")
	_, block16, _ := net.ParseCIDR("192.168.0.0/16")
	_, blockShared, _ := net.ParseCIDR("100.64.0.0/10")
	_, blockLinkLocal, _ := net.ParseCIDR("169.254.0.0/16")
	return block16.Contains(ip) || block20.Contains(ip) || block24.Contains(ip) ||
		blockLinkLocal.Contains(ip) ||
		blockShared.Contains(ip)
}
