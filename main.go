package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"runtime"
	"runtime/debug"
	"slices"
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
		fmt.Println("invalid remote address, exiting...")
		return
	}

	routeV4Cmd := exec.Command("route", "-n", "get", "default")
	resultV4, err := routeV4Cmd.Output()
	if err != nil {
		panic(err)
	}

	routeV6Cmd := exec.Command("route", "-n", "get", "-inet6", "default")
	resultV6, err := routeV6Cmd.Output()
	if err != nil {
		panic(err)
	}

	gwRegex := regexp.MustCompile("gateway: (?P<ip>.*)")
	intfRegex := regexp.MustCompile("interface: (?P<intf>.*)")
	ipRegexIndex := gwRegex.SubexpIndex("ip")
	intfRegexIndex := intfRegex.SubexpIndex("intf")

	defaultV4GatewayBytes, err := getRegexpSubmatch(gwRegex, resultV4, ipRegexIndex)
	if err != nil {
		fmt.Println("invalid IPv4 Gateway...")
		return
	}
	defaultV4Gateway := string(defaultV4GatewayBytes)

	defaultV6Gateway := *new(string)
	defaultV6GatewayBytes, err := getRegexpSubmatch(gwRegex, resultV6, ipRegexIndex)
	if err == nil {
		defaultV6Gateway = string(defaultV6GatewayBytes)
	}

	outboundIntfBytes, err := getRegexpSubmatch(intfRegex, resultV4, intfRegexIndex)
	if err != nil {
		fmt.Println("invalid IPv4 Outbound Interface...")
		return
	}
	outboundIntf := string(outboundIntfBytes)

	uTunName, uTunIndex := createTun()
	cmds := slices.Concat(
		[]string{
			setupInterface(uTunName, utunIPv4),
		},
		replaceDefault(remoteAddress, &utunIPv4, nil, &defaultV4Gateway, false),
	)

	config := &engine.Key{
		Proxy:     inboundProxyURI,
		Device:    fmt.Sprintf("utun%d", uTunIndex),
		Interface: outboundIntf,
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
		execCommand(cmd)
	}

	runtime.GC()
	debug.FreeOSMemory()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	cmds = replaceDefault(remoteAddress, &defaultV4Gateway, &defaultV6Gateway, nil, true)
	for _, cmd := range cmds {
		fmt.Printf("executing: %s\n", cmd)
		execCommand(cmd)
	}
}
