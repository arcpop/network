package main

import (
    "github.com/arcpop/network/ethernet"
    "github.com/arcpop/network/netdev"
    "github.com/arcpop/network/arp"
	"log"
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
}
