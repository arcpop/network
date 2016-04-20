//Package ethernet represents the network stack at layer 2
package ethernet

import (
	"bytes"
	"encoding/binary"
	"github.com/arcpop/network/netdev"
	"log"
	"net"
	"github.com/arcpop/network/config"
)

const (
	//HeaderLength is the length of an ethernet header
	HeaderLength = 14
    
)

var (
	//IPv4In should be defined by ipv4 layer to receive packets
	IPv4In func (*Layer2Packet)
	//IPv6In should be defined by ipv6 layer to receive packets
	IPv6In func (*Layer2Packet)
	//ArpIn should be defined by arp layer to receive packets
	ArpIn func (*Layer2Packet)
	
	BroadcastMACAddress = net.HardwareAddr{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
)

//Layer2Packet represents a Layer 2 packet with ethernet header information
type Layer2Packet struct {
	Dev      netdev.Interface
	L2Header *Header
	Data     []byte
	PacketType int
}

const (
	PacketTypeBroadcast = iota
	PacketTypeMulticast = iota
	PacketTypeUnicast = iota
	PacketTypeLocal = iota
)

//Header represents an ethernet header
type Header struct {
	DstMAC       net.HardwareAddr
	SrcMAC       net.HardwareAddr
	EthernetType uint16
	DataOffset   int
}

//Start starts receiving ethernet frames from the specified NetDev.
func Start(dev netdev.Interface) {
	for i := 0; i < config.Ethernet.NumberOfQueueWorkers; i++ {
		go ethernetRx(dev)
	}
}

func ethHdr(p []byte) *Header {
	if len(p) < 14 {
		return nil
	}
	return &Header{
		DstMAC:       net.HardwareAddr(p[0:6]),
		SrcMAC:       net.HardwareAddr(p[6:12]),
		EthernetType: binary.BigEndian.Uint16(p[12:14]),
		DataOffset:   14,
	}
}

func macAddrCmp(a, b net.HardwareAddr) bool {
	return bytes.Compare(a, b) == 0
}

func ethernetRx(dev netdev.Interface) {
	log.Println("Ethernet: Worker waiting for packets on Interface", dev.GetName())
	for {
		pkt := dev.RxPacket()
		hdr := ethHdr(pkt)
		if hdr == nil {
			log.Println("Ethernet: Malformed ethernet packet (too short)")
			continue
		}
		packet := &Layer2Packet{
			Dev: dev, 
			L2Header: hdr, 
			Data: pkt[hdr.DataOffset:],
			PacketType: PacketTypeUnicast,
		}
		
		if ((hdr.DstMAC[0] & 1) != 0) {
			if bytes.Compare(hdr.DstMAC, BroadcastMACAddress) == 0 {
				packet.PacketType = PacketTypeBroadcast
			} else {
				packet.PacketType = PacketTypeMulticast
				continue //Drop that
			}
		}
		
		
		switch hdr.EthernetType {
		case 0x0800:
			if !macAddrCmp(hdr.DstMAC, dev.GetHardwareAddress()) {
				continue
			}
			IPv4In(packet)
			continue
		case 0x86DD:
			if !macAddrCmp(hdr.DstMAC, dev.GetHardwareAddress()) {
				continue
			}
			IPv6In(packet)
			continue

		case 0x0806:
			ArpIn(packet)
			continue
		default:
			log.Println("Ethernet: Received unclassified packet!")
		}

	}
}
