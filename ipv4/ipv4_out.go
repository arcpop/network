package ipv4

import (
	"encoding/binary"
	"github.com/arcpop/network/ip"
	"github.com/arcpop/network/ethernet"
	"github.com/arcpop/network/util"
	"github.com/arcpop/network/arp"
)


func (h *Header) put(buf []byte) {
    buf[0] = (4 << 4) | 5
    buf[1] = h.TOS
    binary.BigEndian.PutUint16(buf[2:], h.TotalLength)
    binary.BigEndian.PutUint16(buf[4:], h.Identification)
    f := h.FragmentOffset
    if h.DontFragment {
        f |= 0x4000
    }
    if h.MoreFragments {
        f |= 0x2000
    }
    binary.BigEndian.PutUint16(buf[6:], f)
    buf[8] = h.TTL
    buf[9] = h.Protocol
    buf[10] = 0
    buf[11] = 0
    copy(buf[12:16], h.SourceIP)
    copy(buf[16:20], h.TargetIP)
    checksum := ip.InternetChecksum(buf[0:(int(buf[0] & 0xF)) << 2])
    binary.BigEndian.PutUint16(buf[10:], checksum)
}

func AllocatePacket(size int) *L3Packet {
    pkt := make([]byte, size + HeaderLength + ethernet.HeaderLength)
    return &L3Packet{ packetData: pkt, ProtocolData: pkt[HeaderLength + ethernet.HeaderLength:]}
}
func Send(p *L3Packet) error {
    header := p.IPHeader
    pkt := p.packetData
    entry, err := RoutingGetRoute(header.TargetIP)
    if err != nil {
        return err
    }
    header.SourceIP = entry.Iface.GetIPv4Address()
    ProtoData := p.ProtocolData
    mtu := entry.Iface.GetMTU()
    offset := 0
    blockSize := ((mtu - HeaderLength) >> 3) << 3
    header.TotalLength = uint16(len(ProtoData) + HeaderLength)
    //Check if we need to fragment this packet
    if int(header.TotalLength) > mtu {
        if header.DontFragment {
            return ErrPacketTooBig
        }
        for len(ProtoData[offset:]) + HeaderLength > mtu {
            fragHeader := *header
            fragHeader.MoreFragments = true
            fragHeader.FragmentOffset = uint16(offset >> 3)
            fragHeader.TotalLength = uint16(HeaderLength + blockSize)
            p := make([]byte, blockSize + HeaderLength + ethernet.HeaderLength)
            
            fragHeader.put(p[ethernet.HeaderLength:])
            copy(p[ethernet.HeaderLength + HeaderLength:], ProtoData[offset:])
            
            if (entry.flags & FlagGateway) != 0 {
                arp.SetMACAndSend(entry.Iface, p, util.ToIP(entry.gateway))
            } else {
                arp.SetMACAndSend(entry.Iface, p, header.TargetIP)
            }
            offset += blockSize
        }
        rest := len(ProtoData[offset:])
        header.TotalLength = uint16(rest + HeaderLength)
        header.FragmentOffset = uint16(offset >> 3)
        pkt = pkt[0:rest + HeaderLength + ethernet.HeaderLength]
    }
    
    header.put(pkt[ethernet.HeaderLength:])
    copy(pkt[ethernet.HeaderLength + HeaderLength:], ProtoData[offset:])
    if (entry.flags & FlagGateway) != 0 {
        arp.SetMACAndSend(entry.Iface, pkt, util.ToIP(entry.gateway))
    } else {
        arp.SetMACAndSend(entry.Iface, pkt, header.TargetIP)
    }
    
    return nil
}

