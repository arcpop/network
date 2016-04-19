package arp

import (
	"bytes"
	"encoding/binary"
	"github.com/arcpop/network/config"
	"github.com/arcpop/network/ethernet"
	"github.com/arcpop/network/ipv4"
	"github.com/google/gopacket/layers"
	"log"
	"net"
	"sync"
	"time"
)

const (
	//HeaderLength is the arp header length
	HeaderLength = 28
)

type packet struct {
	ethHdr *ethernet.Header
	arpHdr *Header
	dev    *netdev.NetDev
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
	arpCache = make(map[uint32]*arpCacheEntry)
}

//In gets called for each ethernet packet arriving with arp protocol indicator
func In(pkt *ethernet.Layer2Packet) {
	if len(pkt.Data) < HeaderLength {
		log.Prinln("Arp: Packet too short!")
		return
	}
	hdr := parseArpHeader(pkt.Data)
	if bytes.Compare(hdr.targetHWAddr, BroadcastMACAddress) != 0 &&
		bytes.Compare(hdr.targetHWAddr, pkt.Dev.GetHardwareAddress()) != 0 {
		log.Println("Arp: Packet not for us.")
		return
	}
	if hdr.hwAddrType != 1 || hdr.protoAddrType != 0x0800 || hdr.protoAddrLen != 4 || hdr.hwAddrLen != 6 {
		log.Println("Arp: Packet for some other protocol.")
		return
	}
	arpPkt := &Packet{dev: pkt.Dev, ethHdr: pkt.L2Header, arpHdr: hdr}
	go handlePacket(arpPkt)
}

func arpRequest(targetIP net.IP, dev netdev.NetDev) {
	buf := make([]byte, HeaderLength+ethernet.HeaderLength)
	arpPkt := buf[ethernet.HeaderLength:]
	buf[]
}

func handlePacket(arpPkt *packet) {
	if arpPkt.arpHdr.opcode == 1 {
		if arpPkt.arpHdr.targetProtoAddr.IsMulticast() {
			log.Println("Arp: Dropping multicast packet")
			return
		}
		//Request, check if for this device's IP address.
		if bytes.Compare(arpPkt.arpHdr.targetProtoAddr, arpPkt.dev.GetIPv4Addr()) == 0 {
			arpReply()
		}
	}
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
