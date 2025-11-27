//go:build darwin || dragonfly || freebsd || netbsd || openbsd

package main

import (
	"fmt"
	"slices"
)

func deleteDefault() []string {
	return []string{
		"route -n delete -net default",
		"route -n delete -inet6 default",
	}
}

func setupInterface(utunName, ip string) string {
	return fmt.Sprintf(
		`ifconfig '%s' '%s' '%s' up`,
		utunName,
		ip,
		ip,
	)
}

func addIPv4Default(defaultV4Gateway string) string {
	return fmt.Sprintf("route -n add -net default '%s'", defaultV4Gateway)
}

func addIPv6Default(defaultV6Gateway string) string {
	return fmt.Sprintf("route -n add -inet6 -net default '%s'", defaultV6Gateway)
}

func editIP(ip, gw string, delPinning bool) string {
	addHostAction := "add"
	if delPinning {
		addHostAction = "delete"
	}

	return fmt.Sprintf(`route -n '%s' -host '%s' '%s'`, addHostAction, ip, gw)
}

func replaceDefault(
	remoteAddress string,
	defaultV4Gateway, defaultV6Gateway *string,
	internetV4Gateway *string,
	delPinning bool,
) ([]string, error) {
	ipv6Add := []string{}
	if defaultV6Gateway != nil && *defaultV6Gateway != "" {
		ipv6Add = append(ipv6Add, addIPv6Default(*defaultV6Gateway))
	}

	if defaultV4Gateway == nil || *defaultV4Gateway == "" {
		return nil, fmt.Errorf("invalid IPv4 Gateway")
	}

	if internetV4Gateway == nil {
		internetV4Gateway = defaultV4Gateway
	}

	return slices.Concat(
		deleteDefault(),
		[]string{
			addIPv4Default(*defaultV4Gateway),
			editIP(remoteAddress, *internetV4Gateway, delPinning),
		},
		ipv6Add,
	), nil
}
