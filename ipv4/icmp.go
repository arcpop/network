package ipv4

import (
	"encoding/binary"
	"github.com/arcpop/network/ip"
	"log"
	"math/rand"
)


type ICMPPacket struct {
    Type byte
    Code byte
    Checksum uint16
    Data []byte
}

func SendICMPPacket(icmpType, icmpCode byte, header *Header, data []byte)  {
    p := AllocatePacket(len(data) + 4)
    pkt := p.ProtocolData
    p.IPHeader = header
    pkt[0] = icmpType
    pkt[1] = icmpCode
    copy(pkt[4:], data)
    csum := ip.InternetChecksum(pkt)
    binary.BigEndian.PutUint16(pkt[2:4], csum)
    Send(p)
}

func toICMP(pkt []byte) *ICMPPacket {
    csum := ip.InternetChecksum(pkt)
    p := &ICMPPacket{
        Type: pkt[0],
        Code: pkt[1],
        Data: pkt[4:],
    }
    if csum != 0 {
        log.Println("IPv4: ICMP Packet checksum mismatch: ", p.Type, p.Code, csum, binary.BigEndian.Uint16(pkt[2:4]))
        return p
    }
    return p
}

type ICMP struct {
    
}

func (*ICMP) IPv4In(header *Header, pkt[]byte)  {
    icmpPkt := toICMP(pkt)
    if icmpPkt == nil {
        return
    }
    srcIP := header.SourceIP
    switch (icmpPkt.Type) {
    case ip.ICMPTypeEcho:
        if srcIP.IsGlobalUnicast() {
            hdr := &Header{}
            hdr.SourceIP = header.TargetIP
            hdr.TargetIP = srcIP
            hdr.Identification = uint16(rand.Uint32() & 0xFFFF)
            hdr.TTL = 128
            hdr.Protocol = ip.IPPROTO_ICMP
            go SendICMPPacket(ip.ICMPTypeEchoReply, ip.ICMPCodeEchoReply, hdr, icmpPkt.Data)
            return
        }
    default:
        
    }
}

