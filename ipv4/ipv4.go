package ipv4

import (
	"github.com/arcpop/network/ethernet"
	"net"
	"github.com/arcpop/network/ip"
	"errors"
)

const (
    HeaderLength = 20
)
var (
    ErrPacketTooBig = errors.New("Packet needs fragmenting but DontFragment bit is set!")
    ErrInterfaceNotFound = errors.New("Interface not found")
    ErrPacketNotRoutable = errors.New("Packet is not routable!")
)
type L3Packet struct {
    IPHeader *Header
    ProtocolData []byte
    packetData []byte
}
type Header struct {
    headerLength byte
    TOS byte
    TotalLength uint16
    Identification uint16
    FragmentOffset uint16
    DontFragment bool
    MoreFragments bool
    TTL byte
    Protocol byte
    Checksum uint16
    TargetIP net.IP
    SourceIP net.IP
}

func Start()  {
    fragmentationQueue = make(chan *fragment)
    go fragmentationReassemblyWorker()
    ethernet.IPv4In = in
    supportedProtocols = make(map[byte]Protocol)
    initRoutingTable()
    supportedProtocolsLock.Lock()
    supportedProtocols[ip.IPPROTO_ICMP] = &ICMP{}
    supportedProtocolsLock.Unlock()
}