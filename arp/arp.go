package arp

import (
	"net"
	"time"
	"sync"
    "github.com/arcpop/network/ethernet"
    "github.com/arcpop/network/ipv4"
    "github.com/arcpop/network/config"
	"github.com/google/gopacket/layers"
	"log"
	"encoding/binary"
	"bytes"
)

const  (
    MinArpPacketLength = 38
)

type arpCacheEntry struct {
    mac net.HardwareAddr
    validUntil time.Time
}


//arpCache is the arpcache to lookup mac addresses for given ip addresses
arpCache map[uint32] *arpCacheEntry
arpCacheLock sync.RWMutex

//lookupCache contains all sent but not yet received arp requests
lookupCache []*lookup
lookupCacheLock sync.Mutex


type ArpPacket struct {
    ethHdr *ethernet.Header
    arpHdr *Header
}

type Header struct {
    hwAddrType uint16
    protoAddrType uint16
    hwAddrLen byte
    protoAddrLen byte
    opcode uint16
    srcHWAddr net.HardwareAddr
    srcProtoAddr net.IP
    targetHWAddr net.HardwareAddr
    targetProtoAddr net.IP
}

func Start(dev *netdev.NetDev) {
    netDev = dev
    arpCache = make(map[uint32] *arpCacheEntry)
    lookupCache = make([]*lookup)
}

func In(pkt *ethernet.Layer2Packet) {
    hdr := parseArpHeader(pkt.Data)
    if hdr == nil {
        return
    }
    arpPkt := &ArpPacket{ethHdr: pkt.L2Header, arpHdr: hdr}
    go handlePacket(arpPkt)
}
func handlePacket(arpPkt *ArpPacket) {
    if arpPkt.arpHdr.opcode == 1 {
        //Request, check if for this device's IP address.
        if bytes.Compare(arpPkt.arpHdr.targetProtoAddr, netDev.GetIPv4Addr()) == 0 {
            arpReply()
        }
    }
}
func parseArpHeader(pkt []byte) *Header {
    if len(pkt) < 28 {
        log.Prinln("Arp: Packet too short!")
        return nil
    }
    return &Header{
        hwAddrType: binary.BigEndian.Uint16(pkt[0:2]),
        protoAddrType: binary.BigEndian.Uint16(pkt[2:4]),
        hwAddrLen: pkt[4],
        protoAddrLen: pkt[5],
        opcode: binary.BigEndian.Uint16(pkt[6:8]),
        srcHWAddr: pkt[8:14],
        srcProtoAddr: pkt[14:18],
        targetHWAddr: pkt[18:24],
        targetProtoAddr: pkt[24:28],
    }
}
