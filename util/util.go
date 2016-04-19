//Package util contains utility functions needed in the whole project.
package util

import (
	"encoding/binary"
	"net"
	"github.com/arcpop/network/netdev"
)



//ChannelClosed returns true if the stop channel is closed and false otherwise.
func ChannelClosed(stop chan bool) bool {
    select {
        case _, ok := <- stop:
            return ok
    }
}

func IPToUint32(ip []byte) uint32 {
    return binary.BigEndian.Uint32(ip)
}

func ToIP(u uint32) net.IP {
    var buf [4]byte
    binary.BigEndian.PutUint32(buf[:], u)
    return net.IP(buf[:])
}
