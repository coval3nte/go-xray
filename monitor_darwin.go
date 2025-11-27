package main

import (
	"context"
	"fmt"
	"net"
	"slices"

	"golang.org/x/net/route"
	"golang.org/x/sys/unix"
)

func monitor(ctx context.Context, gateway string, cmds []string) error {
	defer ctx.Done()

	fd, err := unix.Socket(unix.AF_ROUTE, unix.SOCK_RAW, 0)
	if err != nil {
		return err
	}

	var buf [2 << 10]byte
	for {
		select {
		case <-ctx.Done():
			fmt.Println("monitor received terminate signal")
			return nil
		default:
		}

		n, err := unix.Read(fd, buf[:])
		if err != nil {
			return err
		}

		msgs, err := route.ParseRIB(route.RIBTypeRoute, buf[:n])
		if err != nil {
			fmt.Println(err)
		}

		for _, msg := range msgs {
			switch msg := msg.(type) {
			case *route.RouteMessage:
				if !slices.Contains(
					[]int{
						unix.RTM_ADD,
						unix.RTM_DELETE,
					},
					msg.Type,
				) {
					continue
				}

				gw := ipOfAddr(msg.Addrs[unix.RTAX_GATEWAY])
				if gw.Equal(nil) {
					continue
				}

				dst := ipOfAddr(msg.Addrs[unix.RTAX_DST])
				if dst.Equal(nil) && !dst.IsUnspecified() {
					continue
				}

				fmt.Printf(
					"type=%s, gateway=%s, dst=%s\n",
					rtmTypeToString(msg.Type), ipOfAddr(msg.Addrs[unix.RTAX_GATEWAY]), ipOfAddr(msg.Addrs[unix.RTAX_DST]),
				)

				if msg.Type == unix.RTM_ADD && !gw.Equal(net.ParseIP(gateway)) && dst.IsUnspecified() {
					for _, cmd := range cmds {
						fmt.Printf("executing: %s\n", cmd)
						execCommand(cmd)
					}
				}
			}
		}
	}
}
