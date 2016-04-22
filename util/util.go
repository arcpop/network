//Package util contains utility functions needed in the whole project.
package util

import (
	"encoding/binary"
	"net"
	"errors"
)

var ErrNotImplemented = errors.New("Functionality not implemented")


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

func Drain(from, to []byte) (fromEmpty bool, n int) {
    needed := len(to)
    available := len(from)
    if available > 0 {
        if available >= needed {
            copy(to[:], from[:needed])
            return false, needed
        }
        copy(to[:available], from[:available])
        needed -= available
    }
    return true, available
}