package net

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var (
	src       = rand.NewSource(time.Now().UnixNano())
	bindLocal = false
	bindIpv6  = false
)

func init() {
	bindLocal, _ = strconv.ParseBool(os.Getenv("BIND_LOCAL"))
	bindIpv6, _ = strconv.ParseBool(os.Getenv("BIND_IPV6"))
}

// RandStringBytesMaskImprSrc returns a random string given a length
func RandStringBytesMaskImprSrc(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

// GetAddresses returns the addresses the transport is listening to
func GetAddresses(l net.Listener) []string {
	port := l.Addr().(*net.TCPAddr).Port
	// TODO log errors
	network := strings.ToLower(l.Addr().Network())
	if network == "tcp" {
		network = "tcps"
	}
	addrs, _ := GetLocalPeerAddresses(port)
	for i, addr := range addrs {
		addrs[i] = fmt.Sprintf("%s:%s", network, addr)
	}
	return addrs
}

func fmtAddress(address string, port int) string {
	return fmt.Sprintf("tcp:%s:%d", address, port)
}

// GetLocalPeerAddresses returns the addresses TCP can listen to on the local machine
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
	if !bindLocal && (ip.IsLoopback() || isPrivate(ip)) {
		return "", false
	}
	if !bindIpv6 && isIPv6(ip.String()) {
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

func isPrivate(ip net.IP) bool {
	_, block24, _ := net.ParseCIDR("10.0.0.0/8")
	_, block20, _ := net.ParseCIDR("172.16.0.0/12")
	_, block16, _ := net.ParseCIDR("192.168.0.0/16")
	return block16.Contains(ip) || block20.Contains(ip) || block24.Contains(ip)
}
