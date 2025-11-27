package main

import "fmt"

var (
	invalidRemote  = fmt.Errorf("invalid remote address")
	invalidTunIPv4 = fmt.Errorf("invalid tun IPv4 address")
)
