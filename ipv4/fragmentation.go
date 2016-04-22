package ipv4


import (
	"log"
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
    completeUntilNow := true
    inserted := false
    for i := 0; i < len(*parts); i++ {
        v := (*parts)[i]
        currentStart := prev.offset + prev.length
        
        //Check if next fragment in slice is not bounding
        if currentStart != v.offset {
            //Check if we could insert the new fragment
            if current.offset >= currentStart {
                //Check if we have enough space for the new fragment
                if (current.offset + current.length) <= v.offset {
                    //Insert new fragment
                    *parts = append(append((*parts)[:i], current), (*parts)[i:]...)
                    //Check if the new fragment is not bounding
                    if (current.offset + current.length) < v.offset {
                        //Not bounding, we are not done yet, there is still some fragment missing
                        return false, false 
                    }
                    //We just check if all remaining fragments are in place
                    prev = current
                    i++
                    inserted = true
                    continue
                } else {
                    return false, true
                }
            } else if !inserted {
                return false, true
            }
        } /* else { //This code is just for having all else clauses, it actually does nothing.
            completeUntilNow = completeUntilNow && true
        } */
        
        prev = v
    }
    return completeUntilNow && prev.lastFragment, !inserted
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

var fragmentationQueue chan *fragment
var fragmentedPackets map[fragmentationKey] []*fragmentationData

func fragmentationReassemblyWorker() {
    for c := range fragmentationQueue {
        parts, ok := fragmentedPackets[c.key]
        if !ok {
            fragmentedPackets[c.key] = []*fragmentationData{c.frag}
        } else {
            complete, collides := checkInsertFragment(c.frag, &parts)
            fragmentedPackets[c.key] = parts
            if collides {
                log.Println("IPv4: Fragment collides with other received fragments, dropping.")
            } else if complete {
                needed := 0
                for _, f := range parts {
                    needed += len(f.data)
                }
                data := make([]byte, needed)
                offset := 0
                for _, f := range parts {
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
    }
}