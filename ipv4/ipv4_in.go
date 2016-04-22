package ipv4

import (
	"github.com/arcpop/network/ethernet"
	"log"
	"sync"
	"github.com/arcpop/network/ip"
	"encoding/binary"
)



func in(pkt *ethernet.Layer2Packet)  {
    hdr := parseHeader(pkt.Data)
    if hdr == nil {
        return
    }
    isFragmented := (hdr.MoreFragments || (hdr.FragmentOffset != 0))
    if hdr.DontFragment && isFragmented {
        log.Println("IPv4: Invalid header fields for fragmentation.")
        return
    }
    headerSize := int(hdr.headerLength) << 2
    protocolData := pkt.Data[headerSize:int(hdr.TotalLength)]
    if isFragmented {
        reassembleFragmented(hdr, protocolData)
        return
    }
    deliverToProtocols(hdr, protocolData)
}

//Send icmp error to all protocols, the responsible one should act upon receiving it
func protocolsCheckForICMPError(hdr *Header, icmpPkt *ICMPPacket) {
    
}

type Protocol interface {
    IPv4In(header *Header, data []byte)
}

var (
    supportedProtocolsLock sync.RWMutex
    supportedProtocols map[byte] Protocol
)

//Here we deliver the ip packets to their corresponding protocol
func deliverToProtocols(hdr *Header, protocolData []byte)  {
    
    if hdr.Protocol == ip.IPPROTO_ICMP {
        dstIP := hdr.TargetIP
        if !dstIP.IsGlobalUnicast() {
            log.Println("IPv4: ICMP Message possibly for broadcast or multicast source, dropping.")
        }
        icmpPkt := toICMP(protocolData)
        if icmpPkt == nil {
            return
        }
        if icmpPkt.Type == ip.ICMPTypeDestinationUnreachable {
            go protocolsCheckForICMPError(hdr, icmpPkt)
            return
        }
    }
    supportedProtocolsLock.RLock()
    proto, ok := supportedProtocols[hdr.Protocol]
    supportedProtocolsLock.RUnlock()
    if !ok {
        log.Println("IPv4: Packet with unsupported protocol: ", hdr.Protocol)
        return
    }
    proto.IPv4In(hdr, protocolData)
}

func parseHeader(buf []byte) *Header {
    if len(buf) < HeaderLength {
        return nil
    }
    version := buf[0] >> 4
    if version != 4 {
        log.Println("IPv4: Invalid version")
        return nil
    }
    csum := ip.InternetChecksum(buf[0:(int(buf[0] & 0xF)) << 2])
    if csum != 0 {
        log.Println("IPv4: Corrupted packet header")
        return nil
    }
    h := &Header{
        headerLength: buf[0] & 0xF,
        TOS: buf[1],
        TotalLength: binary.BigEndian.Uint16(buf[2:4]),
        Identification: binary.BigEndian.Uint16(buf[4:6]),
        FragmentOffset: binary.BigEndian.Uint16(buf[6:8]) & 0x1FFF,
        DontFragment: buf[6] & 0x40 != 0,
        MoreFragments: buf[6] & 0x20 != 0,
        TTL: buf[8],
        Protocol: buf[9],
        Checksum: binary.BigEndian.Uint16(buf[10:12]),
        SourceIP: buf[12:16],
        TargetIP: buf[16:20],
    }
    return h
}