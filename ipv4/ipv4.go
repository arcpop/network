package ipv4

import (
	"github.com/arcpop/network/ethernet"
	"net"
    "encoding/binary"
	"log"
)

const (
    HeaderLength = 20
)

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
    ethernet.IPv4In = in
}

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
    protocolData := pkt.Data[headerSize:int(hdr.TotalLength) - headerSize]
    if isFragmented {
        reassembleFragmented(hdr, protocolData)
        return
    }
    deliverToProtocols(hdr, protocolData)
}
func deliverToProtocols(hdr *Header, protocolData []byte)  {
    
}
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
    var checksum uint32
    i := 0
    length := (int(buf[0] & 0xF)) << 2
    for ; i < length - 1; i += 2 {
        checksum += uint32(buf[i]) << 8
        checksum += uint32(buf[i + 1])
    }
    if i == length - 1 {
        checksum += uint32(buf[i]) << 8
    }
    for carry := (checksum >> 16); carry != 0; carry = (checksum >> 16) {
        checksum = (checksum & 0xFFFF) + carry
    }
    binary.BigEndian.PutUint16(buf[10:], uint16(^checksum))
}

func checksum(buf []byte) uint16 {
    length := (int(buf[0] & 0xF)) << 2
    var csum uint32
    i := 0
    for ; i < length - 1; i += 2 {
        csum += uint32(buf[i]) << 8
        csum += uint32(buf[i + 1])
    }
    if i == length - 1 {
        csum += uint32(buf[i]) << 8
    }
    for carry := (csum >> 16); carry != 0; carry = (csum >> 16) {
        csum = (csum & 0xFFFF) + carry
    }
    return uint16(^csum) 
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
    csum := checksum(buf)
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
        TargetIP: buf[16:20],
        SourceIP: buf[12:16],
    }
    return h
}