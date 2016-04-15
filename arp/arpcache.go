package arp

import (
    "net"
	"encoding/binary"
	"sync"
	"github.com/google/gopacket/layers"
	"github.com/arcpop/network/ethernet"
)

var (
    //BroadcastMACAddress is the broadcast hw address to send arp requests to.
    BroadcastMACAddress = net.HardwareAddr{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
)

type lookup struct {
    ip uint32
    lock sync.Mutex
    waiters []chan bool
}

type arpQuery struct {
    hwType uint16
    protocolType layers.EthernetType
    hwSize byte
    protocolSize byte
    opcode uint16
    senderMAC net.HardwareAddr
    senderIP net.IP
    targetMAC net.HardwareAddr
    targetIP net.IP
}

//FindInCache finds an address in cache or gets it via an arp request
func (al *ArpLayer) FindInCache(ip net.IP) (net.HardwareAddr, error) {
    return nil, nil
}

//Lookup performs an arp query
func (al *ArpLayer) Lookup(targetIP net.IP) (net.HardwareAddr, error) {
    var waiter chan bool
    
    //Check if already looking up
    ip32 := binary.BigEndian.Uint32(targetIP.To4())
    al.lookupCacheLock.Lock()
    for _, i := range al.lookupCache {
        i.lock.Lock()
        if i.ip == ip32 {
            //Someone already requested a lookup, we queue into the waiters queue
            waiter = make(chan bool)
            i.waiters = append(i.waiters, waiter)
            i.lock.Unlock()
            break
        }
        i.lock.Unlock()
    }
    localIP := al.ip4.GetLocalAddressForSubnet(targetIP)
    if localIP == nil {
        al.lookupCacheLock.Unlock()
        return nil, ErrInvalidAddressRequested
    }
    //We are the first, thus we must also send the arp request
    if waiter == nil {
        waiter = make(chan bool)
        l := &lookup{ip: ip32, waiters: []chan bool{ waiter, }}
        al.lookupCache = append(al.lookupCache, l)
        q := &arpQuery{
            hwType: 1,
            protocolType: layers.EthernetTypeIPv4,
            hwSize: 6,
            protocolSize: 4,
            senderMAC: al.el.GetMACAddress(),
            senderIP: senderIP,
            targetMAC: make([]byte, 6),
            targetIP: targetIP,
        }
        pkt := encode(q)
        al.senderQueue <- &ethernet.Layer3Paket{Data: pkt, DstMAC: BroadcastMACAddress}
    }
    al.lookupCacheLock.Unlock()
    
    //We block until either the channel is closed or we got a message.
    //The worker automatically puts the replies into the cache and notifies all channels who looked it up.
    _, _ = <-waiter
    
    return al.FindInCache(targetIP)
}

func encode(q *arpQuery) []byte  {
    p := make([]byte, 60)
    binary.BigEndian.PutUint16(p[0:2], q.hwType)
    binary.BigEndian.PutUint16(p[2:4], uint16(q.protocolType))
    p[4] = q.hwSize
    p[5] = q.protocolSize
    binary.BigEndian.PutUint16(p[6:8], q.opcode)
    copy(p[8:14], q.senderMAC)
    copy(p[14:18], q.senderIP)
    copy(p[18:24], q.targetMAC)
    copy(p[24:28], q.targetIP)
    return p
}