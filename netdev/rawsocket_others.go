// +build !linux

package netdev

import (
    "github.com/arcpop/network/util"
)

func NewRawSocket(ifname string) (Interface, error) {
    return nil, util.ErrNotImplemented
}
