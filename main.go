package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"slices"
	"sync"
	"syscall"

	"github.com/xjasonlyu/tun2socks/v2/engine"
	"github.com/xtls/xray-core/core"
	_ "github.com/xtls/xray-core/main/distro/all"
)

func main() {
	remoteAddressFlag := flag.String("remote-address", "", "remote IPv4 Address")
	configFileFlag := flag.String("config-file", "config.json", "JSON encoded XRay Config")
	utunIPv4Flag := flag.String("utun-v4", "192.168.18.1", "utun ephemeral IPv4")
	inboundProxyURIFlag := flag.String("inbound-proxy", "socks5://127.0.0.1:1080", "Xray Inbound Proxy")
	flag.Parse()

	remoteAddress := *remoteAddressFlag
	configFile := *configFileFlag
	utunIPv4 := *utunIPv4Flag
	inboundProxyURI := *inboundProxyURIFlag

	if remoteAddress == "" {
		flag.PrintDefaults()
		fmt.Println("\nerror: invalid remote address")
		return
	}

	ipv4Gateway, err := getDefaultGateway(false)
	if err != nil {
		panic(err)
	}

	ipv6Gateway, _ := getDefaultGateway(true)

	replaceString, err := replaceDefault(remoteAddress, &utunIPv4, nil, &ipv4Gateway.Gateway, false)
	if err != nil {
		panic(err)
	}

	uTunName, _ := createTun()
	fmt.Printf("found available tun interface: %s\n", uTunName)

	cmds := slices.Concat(
		[]string{
			setupInterface(uTunName, utunIPv4),
		},
		replaceString,
	)

	config := &engine.Key{
		Proxy:     inboundProxyURI,
		Device:    uTunName,
		Interface: ipv4Gateway.Interface,
		LogLevel:  "warn",
	}
	engine.Insert(config)
	engine.Start()
	defer engine.Stop()

	xRayConfigBytes, err := os.ReadFile(configFile)
	if err != nil {
		panic(err)
	}

	xrayInstance, err := core.StartInstance("json", xRayConfigBytes)
	if err != nil {
		panic(err)
	}

	if err := xrayInstance.Start(); err != nil {
		panic(err)
	}

	defer xrayInstance.Close()

	for _, cmd := range cmds {
		fmt.Printf("executing: %s\n", cmd)
		retryBackOff(func() error {
			_, err := execCommand(cmd)
			if err != nil {
				fmt.Printf("command failed: %v\n", err)
			}

			return err
		}, 3)
	}

	sigCh := make(chan os.Signal, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := *new(sync.WaitGroup)
	wg.Go(func() {
		if err := monitor(ctx, utunIPv4, replaceString); err != nil {
			fmt.Printf("monitor err: %v\n", err)
			sigCh <- os.Kill
		}
	})

	runtime.GC()
	debug.FreeOSMemory()

	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	cancel()
	wg.Wait()

	cmds, err = replaceDefault(remoteAddress, &ipv4Gateway.Gateway, &ipv6Gateway.Gateway, nil, true)
	if err != nil {
		panic(err)
	}

	for _, cmd := range cmds {
		fmt.Printf("executing: %s\n", cmd)
		retryBackOff(func() error {
			_, err := execCommand(cmd)
			if err != nil {
				fmt.Printf("command failed: %v\n", err)
			}

			return err
		}, 3)
	}
}
