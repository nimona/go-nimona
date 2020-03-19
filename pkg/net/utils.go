package net

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

var (
	BindLocal   = false // TODO(geoah) refactor to remove global
	BindPrivate = false // TODO(geoah) refactor to remove global
	bindIpv6    = false // TODO(geoah) refactor to remove global
)

// TODO remove Binds and replace with options
// nolint: gochecknoinits
func init() {
	BindLocal, _ = strconv.ParseBool(os.Getenv("BIND_LOCAL"))
	BindPrivate, _ = strconv.ParseBool(os.Getenv("BIND_PRIVATE"))
	bindIpv6, _ = strconv.ParseBool(os.Getenv("BIND_IPV6"))
}

// GetAddresses returns the addresses the transport is listening to
func GetAddresses(protocol string, l net.Listener) []string {
	port := l.Addr().(*net.TCPAddr).Port
	// TODO log errors
	addrs, _ := GetLocalPeerAddresses(port)
	for i, addr := range addrs {
		addrs[i] = fmt.Sprintf("%s:%s", protocol, addr)
	}
	return addrs
}

func fmtAddress(protocol, address string, port int) string {
	return fmt.Sprintf("%s:%s:%d", protocol, address, port)
}

// GetLocalPeerAddresses returns the addresses TCP can listen to on the
// local machine
func GetLocalPeerAddresses(port int) ([]string, error) {
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

func isValidIP(addr net.Addr) (string, bool) {
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
	if !BindLocal && ip.IsLoopback() {
		return "", false
	}
	if !BindPrivate && isPrivate(ip) {
		return "", false
	}
	if !bindIpv6 && isIPv6(ip.String()) {
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
	_, blockLinkLocal, _ := net.ParseCIDR("169.254.0.0/16")
	return block16.Contains(ip) || block20.Contains(ip) || block24.Contains(ip) ||
		blockLinkLocal.Contains(ip)
}
