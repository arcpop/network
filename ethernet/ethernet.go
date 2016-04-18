//Package ethernet represents the network stack at layer 2
package ethernet

import (
	"encoding/binary"
	"errors"
	"github.com/arcpop/network/netdev"
	"net"
	"bytes"
	"github.com/arcpop/network/config"
	"log"
	"github.com/arcpop/network/ipv4"
)

//Layer2Packet represents a Layer 2 packet with ethernet header information
type Layer2Packet struct {
	Dev *netdev.NetDev
	L2Header *Header
	Data []byte
}

//Header represents an ethernet header
type Header struct {
	DstMAC       net.HardwareAddr
	SrcMAC       net.HardwareAddr
	EthernetType uint16
	DataOffset   int
}

//Start starts receiving ethernet frames from the specified NetDev.
func Start(dev *netdev.NetDev) {
	for i := 0; i < config.Ethernet.NumberOfQueueWorkers; i++ {
		go ethernetRx(dev)
	}
}

func ethHdr(p []byte) *Header {
	if len(p) < 14 {
		return nil
	}
	return &Header{
		DstMAC: net.HardwareAddr(p[0:6]),
		SrcMAC: net.HardwareAddr(p[6:12]),
		EthernetType: binary.BigEndian.Uint16(p[12:14]), 
		DataOffset: 14,
	}
}

func macAddrCmp(a, b net.HardwareAddr) bool {
	return bytes.Compare(a, b) == 0
}

func ethernetRx(dev *netdev.NetDev) {
	for {
		pkt := dev.RxPacket(dev)
		hdr := ethHdr(pkt)
		if hdr == nil {
			log.Println("Ethernet: Malformed ethernet packet (too short)")
			continue
		}
		//Check if multicast packet
		if (hdr.DstMAC[0] & 1) != 0 {
			log.Println("Ethernet: Received multicast packet...")
			continue
		}
		packet := &Layer2Packet{ Dev: dev, L2hdr: hdr, Data: pkt[hdr.DataOffset:]}
		switch (hdr.EthernetType) {
		case 0x0800:
			if !macAddrCmp(hdr.DstMAC, config.Ethernet.MACAddress) {
				log.Println("Ethernet: IPv4 Packet with wrong MAC address")
				continue
			} 
			ipv4.In(packet)
			continue
		case 0x86DD:
			if !macAddrCmp(hdr.DstMAC, config.Ethernet.MACAddress) {
				log.Println("Ethernet: IPv6 Packet with wrong MAC address")
				continue
			} 
			ipv6.In(packet)
			continue
		
		case 0x0806:
			arp.In(packet)
			continue
		default:
			log.Println("Ethernet: Received unclassified packet!")
		}
		
	}
}

