// Package resolver implements a public IP resolver
package resolver

import (
	"net"
	"os"
)

// ResolveHostIP is a helper function that will recover the hosts public IP address or default to 127.0.0.1
func ResolveHostIP() (addr string, err error) {
	addr = "127.0.0.1"
	name, err := os.Hostname()
	if err != nil {
		return addr, err
	}
	ips, err := net.LookupIP(name)
	if err != nil {
		return addr, err
	}
	for _, ip := range ips {
		if !(ip.IsLoopback()) {
			return ip.String(), err
		}
	}
	return addr, err
}
