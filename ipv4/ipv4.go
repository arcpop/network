package ipv4

import (
	"net"
)

type IPv4Layer struct {
    
}


//GetLocalAddressForSubnet returns one of the local addresses which is in the same subnet as otherAddress.
//This is used for arp requests.
func (ip *IPv4Layer) GetLocalAddressForSubnet(otherAddress net.IP) (net.IP) {
    return nil
}
