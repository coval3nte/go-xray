//go:build darwin || dragonfly || freebsd || netbsd || openbsd

package main

import (
	"context"
	"fmt"
	"net/netip"
	"slices"

	"golang.org/x/net/route"
	"golang.org/x/sys/unix"
)

var routes = make(chan *Route)

func monitor(ctx context.Context, config *Config) error {
	defer ctx.Done()

	fd, err := unix.Socket(unix.AF_ROUTE, unix.SOCK_RAW, 0)
	if err != nil {
		return err
	}

	var buf [2 << 10]byte
	routineError := make(chan error, 1)

	go func() {
		for {
			n, err := unix.Read(fd, buf[:])
			if err != nil {
				routineError <- err
				return
			}

			msgs, err := route.ParseRIB(route.RIBTypeRoute, buf[:n])
			if err != nil {
				fmt.Printf("[monitor] parseRIB error: %v\n", err)
			}

			for _, msg := range msgs {
				switch msg := msg.(type) {
				case *route.RouteMessage:
					handleRouteMessage(config, msg)
				}
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("[monitor] received terminate signal")
			return nil
		case err := <-routineError:
			return err
		case route := <-routes:
			config.internetIPv4Gateway.Gateway = route.Gateway

			cmds, err := replaceDefault(
				config.remoteAddress,
				&config.utunIPv4,
				nil,
				&config.internetIPv4Gateway.Gateway,
				false,
			)
			if err != nil {
				fmt.Printf("[monitor] replaceDefault error: %v\n", err)
			}

			cmds = append([]string{editIP(config.remoteAddress, config.internetIPv4Gateway.Gateway, true)}, cmds...)
			for _, cmd := range cmds {
				fmt.Printf("[monitor] executing: %s\n", cmd)
				execCommand(cmd)
			}
		}
	}
}

func handleRouteMessage(config *Config, msg *route.RouteMessage) {
	if !slices.Contains(
		[]int{
			unix.RTM_ADD,
			unix.RTM_DELETE,
			// unix.RTM_CHANGE,
		},
		msg.Type,
	) {
		return
	}

	gw := ipOfAddr(msg.Addrs[unix.RTAX_GATEWAY])
	if !gw.IsValid() || gw.IsUnspecified() {
		return
	}

	dst := ipOfAddr(msg.Addrs[unix.RTAX_DST])
	if !dst.IsValid() || !dst.IsUnspecified() {
		return
	}

	fmt.Printf(
		"[monitor] type=%s, gateway=%s, dst=%s\n",
		rtmTypeToString(msg.Type), ipOfAddr(msg.Addrs[unix.RTAX_GATEWAY]), ipOfAddr(msg.Addrs[unix.RTAX_DST]),
	)

	if msg.Type == unix.RTM_ADD &&
		((gw != netip.MustParseAddr(config.utunIPv4)) || (gw.Is6())) {
		if gw.Is6() {
			routes <- &Route{
				Gateway: config.internetIPv4Gateway.Gateway,
			}
			return
		}

		routes <- &Route{
			Gateway: gw.String(),
		}
	}
}
