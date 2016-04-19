package ipv4

import (
	"net"
    "sync"
	"github.com/arcpop/network/netdev"
)




type routingEntry struct {
    netmask net.IP
    network net.IP
    gateway net.IP
    routeDev* netdev.NetDev
}

var (
    routingTable []*routingEntry
    routingTableLock sync.RWMutex
)

func routingSendPacket(pkt []byte, targetIP net.IP) (error) {
    
}


func initRoutingTable() {
    
}