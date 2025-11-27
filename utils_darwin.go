//go:build darwin || dragonfly || freebsd || netbsd || openbsd

package main

import (
	"net"

	"golang.org/x/net/route"
	"golang.org/x/sys/unix"
)

func ipOfAddr(a route.Addr) net.IP {
	switch a := a.(type) {
	case *route.Inet4Addr:
		return net.IPv4(a.IP[0], a.IP[1], a.IP[2], a.IP[3])
	case *route.Inet6Addr:
		ip := net.IP(a.IP[:])
		return ip
	}
	return net.IP{}
}

func rtmTypeToString(t int) string {
	switch t {
	case unix.RTM_ADD:
		return "ADD"
	case unix.RTM_DELETE:
		return "DELETE"
	default:
		return ""
	}
}
