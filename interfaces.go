package main

import (
	"fmt"
	"net"
	"strings"
)

func createTun() (string, int) {
	intfs, err := net.Interfaces()
	if err != nil {
		panic(err)
	}

	intfInt := -1
	intfGap := false
	for _, intf := range intfs {
		if strings.HasPrefix(intf.Name, "utun") {
			previous := intfInt
			fmt.Sscanf(intf.Name, "utun%d", &intfInt)

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

	return fmt.Sprintf("utun%d", intfInt), intfInt
}
