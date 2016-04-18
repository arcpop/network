package ipv4

import (
	"net"
	"github.com/arcpop/network/ethernet"
)


func In(pkt *ethernet.Layer2Packet)  {
    
}

//GetLocalAddressForSubnet returns one of the local addresses which is in the same subnet as otherAddress.
//This is used for arp requests.
func GetLocalAddressForSubnet(otherAddress net.IP) (net.IP) {
    return nil
}

func IsLocalIP(addr net.IP) bool {
    return false
}