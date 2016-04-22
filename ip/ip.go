package ip


const (
    IPPROTO_ICMP = 1
    IPPROTO_TCP = 6
    IPPROTO_UDP = 17
)
const (
    ICMPTypeEchoReply = 0
    ICMPCodeEchoReply = 0
    
    ICMPTypeDestinationUnreachable = 3
    ICMPCodeNetUnreachable = 1
    ICMPCodeHostUnreachable = 2
    ICMPCodeProtocolUnreachable = 3
    ICMPCodeFragmentationNeeded = 4
    ICMPCodeSourceRouteFailed = 5
    ICMPCodeDestinationNetworkUnknown = 6
    ICMPCodeDestinationHostUnknown = 7
    ICMPCodeSourceHostIsolated = 8
    
    ICMPTypeRedirect = 5
    ICMPCodeRedirectNetwork = 0
    ICMPCodeRedirectHost = 1
    ICMPCodeRedirectTypeOfServiceAndNetwork = 2
    ICMPCodeRedirectTypeOfServiceAndHost = 3
    
    ICMPTypeEcho = 8
    ICMPCodeEcho = 0
    
    //ICMPv6
    ICMPTypeRouterAdvertisement = 9
    ICMPTypeRouterSolicitation = 10
    
    ICMPTypeTimeExceeded = 11
    ICMPCodeTTLExceededInTransmit = 0
    ICMPCodeFragmentReassemblyTimeout = 1
    
    ICMPTypeParameterProblem = 12
    ICMPCodePointerIndicatesProblem = 0
    ICMPCodeMissingRequiredOption = 1
    ICMPCodeBadLength = 2
    
    ICMPTypeTimestamp = 13
    ICMPCodeTimestamp = 0
    
    ICMPTypeTimestampReply = 14
    ICMPCodeTimestampReply = 0
    
)

func InternetChecksum(pkt []byte) uint16 {
    var csum uint32
    i := 0;
    for  ; i < len(pkt) - 1; i += 2 {
        csum += uint32(pkt[i]) << 8
        csum += uint32(pkt[i + 1])
    }
    if i == len(pkt) - 1 {
        csum += uint32(pkt[i]) << 8
    }
    for carry := (csum >> 16); carry != 0; carry = (csum >> 16) {
        csum = (csum & 0xFFFF) + carry
    }
    return ^uint16(csum)
}