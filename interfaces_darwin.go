//go:build darwin || dragonfly || freebsd || netbsd || openbsd

package main

import (
	"fmt"
	"net"
	"regexp"
	"strings"
)

var (
	gwRegex   = regexp.MustCompile("gateway: (?P<ip>.*)")
	intfRegex = regexp.MustCompile("interface: (?P<intf>.*)")

	ipRegexIndex   = gwRegex.SubexpIndex("ip")
	intfRegexIndex = intfRegex.SubexpIndex("intf")

	tunPrefix string = "utun"
)

type Route struct {
	Gateway   string
	Interface string
}

func getDefaultGateway(isV6 bool) (*Route, error) {
	cmd := []string{"route", "-n", "get"}
	if isV6 {
		cmd = append(cmd, "-inet6")
	}
	cmd = append(cmd, "default")

	result, err := execCommand(strings.Join(cmd, " "))
	if err != nil {
		panic(err)
	}

	defaultGatewayBytes, err := getRegexpSubmatch(gwRegex, result, ipRegexIndex)
	if err != nil {
		return nil, fmt.Errorf("invalid Gateway...")
	}

	outboundIntfBytes, err := getRegexpSubmatch(intfRegex, result, intfRegexIndex)
	if err != nil {
		return nil, fmt.Errorf("invalid Outbound Interface...")
	}

	return &Route{
		Gateway:   string(defaultGatewayBytes),
		Interface: string(outboundIntfBytes),
	}, nil
}

func createTun() (string, int) {
	intfs, err := net.Interfaces()
	if err != nil {
		panic(err)
	}

	intfInt := -1
	intfGap := false
	for _, intf := range intfs {
		if strings.HasPrefix(intf.Name, tunPrefix) {
			previous := intfInt
			fmt.Sscanf(intf.Name, tunPrefix+"%d", &intfInt)

			if intfInt > previous+1 {
				intfGap = true
				intfInt = previous + 1
				break
			}
		}
	}

	if intfInt == -1 {
		intfInt = 0
	} else if !intfGap {
		intfInt++
	}

	return getTunName(intfInt), intfInt
}

func getTunName(index int) string {
	return fmt.Sprintf("%s%d", tunPrefix, index)
}
