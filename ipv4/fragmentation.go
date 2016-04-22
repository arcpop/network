package ipv4


import (
	"log"
	"time"
)

type fragmentationKey struct {
    dstIP [4]byte
    srcIP [4]byte
    id uint16
    protocol byte
}

type fragmentationData struct {
    offset uint32
    length uint32
    data []byte
    lastFragment bool
}


func checkInsertFragment(current *fragmentationData, parts *[]*fragmentationData) (complete, collides bool) {
    prev := &fragmentationData {}
    inserted := false
    for i := 0; i < len(*parts); i++ {
        
        v := (*parts)[i]
        
        currentStart := prev.offset + prev.length
        
        
        if !inserted {
            //Current matches at this position
            if current.offset == currentStart {
                if v.offset < current.offset + current.length {
                    return false, true
                }
                after := (*parts)[i:]
                *parts = append(append((*parts)[:i], current), after...)
                inserted = true
                prev = current
                i++
                continue
            }
        } else {
            //Check if complete
            if v.offset > currentStart {
                return false, false
            }
        }
        prev = v
    }
    
    currentStart := prev.offset + prev.length
    if current.offset >= currentStart {
        *parts = append(*parts, current)
        inserted = true
        prev = current
    }
    
    return prev.lastFragment, !inserted
}

func reassembleFragmented(hdr *Header, protocolData []byte) {
    k := fragmentationKey {
        id: hdr.Identification,
        protocol: hdr.Protocol,
    }
    copy(k.dstIP[:], hdr.TargetIP)
    copy(k.srcIP[:], hdr.SourceIP)
    frag := &fragmentationData{
        offset: uint32(hdr.FragmentOffset) << 3,
        length: uint32(len(protocolData)),
        data: protocolData,
        lastFragment: !hdr.MoreFragments,
    }
    fragmentationQueue <- &fragment{key: k, frag: frag}
}

type fragment struct {
    key fragmentationKey
    frag *fragmentationData
}
type fragmentMapEntry struct {
    parts []*fragmentationData
    lastUpdated time.Time
}

var fragmentationQueue chan *fragment
var fragmentedPackets map[fragmentationKey] *fragmentMapEntry

func fragmentationReassemblyWorker() {
    ticker := time.NewTicker(time.Second * 5)
    fragmentedPackets = make(map[fragmentationKey] *fragmentMapEntry)
    for {
        select {
        case c := <- fragmentationQueue:
            parts, ok := fragmentedPackets[c.key]
            if !ok {
                e := &fragmentMapEntry{
                    parts: []*fragmentationData{c.frag},
                    lastUpdated: time.Now(),
                }
                fragmentedPackets[c.key] = e
            } else {
                complete, collides := checkInsertFragment(c.frag, &(parts.parts))
                fragmentedPackets[c.key] = parts
                if collides {
                    log.Println("IPv4: Fragment collides with other received fragments, dropping.")
                } else if complete {
                    needed := 0
                    for _, f := range parts.parts {
                        needed += len(f.data)
                    }
                    data := make([]byte, needed)
                    offset := 0
                    for _, f := range parts.parts {
                        copy(data[offset:], f.data)
                        offset += len(f.data)
                    }
                    delete(fragmentedPackets, c.key)
                    hdr := &Header{
                        SourceIP: c.key.srcIP[:],
                        TargetIP: c.key.dstIP[:],
                        TotalLength: 20 + uint16(offset),
                        Identification: c.key.id,
                        Protocol: c.key.protocol,
                    }
                    go deliverToProtocols(hdr, data)
                }
            }
        case _ = <- ticker.C:
            lastAllowedTime := time.Now().Add(-1 * time.Minute)
            for k,v := range fragmentedPackets {
                if v.lastUpdated.Before(lastAllowedTime) {
                    delete(fragmentedPackets, k)
                }
            }
        }
    }
}