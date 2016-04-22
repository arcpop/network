package main

import (
    "github.com/arcpop/network/ethernet"
    "github.com/arcpop/network/netdev"
    "github.com/arcpop/network/arp"
    "github.com/arcpop/network/ipv4"
    "github.com/arcpop/network/shell"
	"log"
	"net"
)

func main() {
    defer netdev.ShutdownInterfaces()
	loopback, err := netdev.NewLoopback("lo")
    if err != nil {
        log.Println(err)
        return
    }
    eth1, err := netdev.NewRawSocket("eth1")
    if err != nil {
        log.Println(err)
        return
    }
    arp.Start()
    ethernet.Start(loopback)
    ethernet.Start(eth1)
    ipv4.Start()
    ipv4.ConfigureInterfaceAddress("eth1", net.IPNet{ IP:net.IP{192, 168, 56, 101}, Mask: net.IPMask{255, 255, 255, 0}})
    shell.Run()
}
