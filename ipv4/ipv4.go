package ipv4

import (
	"github.com/arcpop/network/ethernet"
)

func Start()  {
    ethernet.IPv4In = in
}

func in(pkt *ethernet.Layer2Packet)  {
    
}
