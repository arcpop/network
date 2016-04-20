package arp

import (
	"bytes"
	"encoding/binary"
	"github.com/arcpop/network/ethernet"
	"github.com/arcpop/network/netdev"
	"log"
	"net"
)

const (
	//HeaderLength is the arp header length
	HeaderLength = 28
)

type packet struct {
	ethHdr *ethernet.Header
	arpHdr *Header
	dev    netdev.Interface
}

//Header represents an arp header
type Header struct {
	hwAddrType      uint16
	protoAddrType   uint16
	hwAddrLen       byte
	protoAddrLen    byte
	opcode          uint16
	srcHWAddr       net.HardwareAddr
	srcProtoAddr    net.IP
	targetHWAddr    net.HardwareAddr
	targetProtoAddr net.IP
}

//Start starts the arp layer
func Start() {
	ethernet.ArpIn = in
	arpCache = make(map[uint32]*arpCacheEntry)
}

func in(pkt *ethernet.Layer2Packet) {
	if len(pkt.Data) < HeaderLength {
		log.Println("Arp: Packet too short!")
		return
	}
	hdr := parseArpHeader(pkt.Data)
	if bytes.Compare(hdr.srcHWAddr, pkt.L2Header.SrcMAC) != 0 {
		log.Println("Arp: Dropping possible spoofed packet!")
		return
	}
	
	if bytes.Compare(hdr.targetHWAddr, BroadcastMACAddress) != 0 &&
		bytes.Compare(hdr.targetHWAddr, pkt.Dev.GetHardwareAddress()) != 0 {
		log.Println("Arp: Packet not for us.")
		return
	}
	if hdr.hwAddrType != 1 || hdr.protoAddrType != 0x0800 || hdr.protoAddrLen != 4 || hdr.hwAddrLen != 6 {
		log.Println("Arp: Packet for some other protocol.")
		return
	}
	arpPkt := &packet{dev: pkt.Dev, ethHdr: pkt.L2Header, arpHdr: hdr}
	go handlePacket(arpPkt)
}

func arpRequest(targetIP net.IP, dev netdev.Interface) {
	buf := make([]byte, HeaderLength+ethernet.HeaderLength)
	copy(buf[0:6], BroadcastMACAddress)
	copy(buf[6:12], dev.GetHardwareAddress())
	binary.BigEndian.PutUint16(buf[12:14], 0x0806)
	arpPkt := buf[ethernet.HeaderLength:]
	binary.BigEndian.PutUint16(arpPkt[0:2], 0x0001)
	binary.BigEndian.PutUint16(arpPkt[2:4], 0x0800)
	arpPkt[4] = 6
	arpPkt[5] = 4
	binary.BigEndian.PutUint16(arpPkt[6:8], 1)
	copy(arpPkt[8:14], dev.GetHardwareAddress())
	copy(arpPkt[14:18], dev.GetIPv4Address())
	copy(arpPkt[18:24], BroadcastMACAddress)
	copy(arpPkt[24:28], targetIP)
	dev.TxPacket(buf)
}

func arpReply(targetIP net.IP, targetMAC net.HardwareAddr,dev netdev.Interface) {
	buf := make([]byte, HeaderLength+ethernet.HeaderLength)
	copy(buf[0:6], targetMAC)
	copy(buf[6:12], dev.GetHardwareAddress())
	binary.BigEndian.PutUint16(buf[12:14], 0x0806)
	arpPkt := buf[ethernet.HeaderLength:]
	binary.BigEndian.PutUint16(arpPkt[0:2], 0x0002)
	binary.BigEndian.PutUint16(arpPkt[2:4], 0x0800)
	arpPkt[4] = 6
	arpPkt[5] = 4
	binary.BigEndian.PutUint16(arpPkt[6:8], 1)
	copy(arpPkt[8:14], dev.GetHardwareAddress())
	copy(arpPkt[14:18], dev.GetIPv4Address())
	copy(arpPkt[18:24], targetMAC)
	copy(arpPkt[24:28], targetIP)
	dev.TxPacket(buf)
}

func handlePacket(arpPkt *packet) {
	if arpPkt.arpHdr.opcode == 1 {
		if arpPkt.arpHdr.targetProtoAddr.IsMulticast() {
			log.Println("Arp: Dropping multicast packet")
			return
		}
		//Request, check if for this device's IP address.
		if bytes.Compare(arpPkt.arpHdr.targetProtoAddr, arpPkt.dev.GetIPv4Address()) == 0 {
			arpReply(arpPkt.arpHdr.srcProtoAddr, arpPkt.arpHdr.srcHWAddr, arpPkt.dev)
		} 
		
	} 
	//We update also from arp requests since this improves our cache
	cacheUpdate(arpPkt.dev, arpPkt.arpHdr.srcProtoAddr, arpPkt.arpHdr.srcHWAddr)
}

func parseArpHeader(pkt []byte) *Header {
	return &Header{
		hwAddrType:      binary.BigEndian.Uint16(pkt[0:2]),
		protoAddrType:   binary.BigEndian.Uint16(pkt[2:4]),
		hwAddrLen:       pkt[4],
		protoAddrLen:    pkt[5],
		opcode:          binary.BigEndian.Uint16(pkt[6:8]),
		srcHWAddr:       pkt[8:14],
		srcProtoAddr:    pkt[14:18],
		targetHWAddr:    pkt[18:24],
		targetProtoAddr: pkt[24:28],
	}
}
