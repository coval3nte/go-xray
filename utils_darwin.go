//go:build darwin || dragonfly || freebsd || netbsd || openbsd

package main

import (
	"net/netip"

	"golang.org/x/net/route"
	"golang.org/x/sys/unix"
)

func ipOfAddr(a route.Addr) netip.Addr {
	switch a := a.(type) {
	case *route.Inet4Addr:
		return netip.AddrFrom4(a.IP)
	case *route.Inet6Addr:
		return netip.AddrFrom16(a.IP)
	}

	return netip.Addr{}
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
