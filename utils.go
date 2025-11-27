package main

import (
	"fmt"
	"net"
	"os/exec"
	"regexp"

	"github.com/google/shlex"
	"golang.org/x/net/route"
	"golang.org/x/sys/unix"
)

func retryBackOff(t func() error, retries int) {
	for range retries {
		if err := t(); err == nil {
			return
		}
	}
}

func execCommand(cmd string) ([]byte, error) {
	parts, err := shlex.Split(cmd)
	if err != nil {
		return nil, err
	}

	if len(parts) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	return exec.Command(parts[0], parts[1:]...).Output()
}

func getRegexpSubmatch(re *regexp.Regexp, b []byte, index int) ([]byte, error) {
	if re == nil {
		return nil, fmt.Errorf("invalid regex")
	} else if matches := re.FindSubmatch(b); len(matches)-1 >= index {
		return matches[index], nil
	}

	return nil, fmt.Errorf("no matches for the supplied capture group found")
}

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
