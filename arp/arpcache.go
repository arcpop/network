package arp

import (
    "net"
	"encoding/binary"
	"sync"
	"github.com/google/gopacket/layers"
	"github.com/arcpop/network/ethernet"
	"errors"
	"time"
)

var (
    //BroadcastMACAddress is the broadcast hw address to send arp requests to.
    BroadcastMACAddress = net.HardwareAddr{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
)

var ErrInvalidAddressRequested = errors.New("Invalid target in arp request with no local ip address")


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

func findInCache(ip net.IP) (net.HardwareAddr, error) {
    ip32 := binary.BigEndian.Uint32(ip)
    arpCacheLock.RLock()
    entry, ok := al.arpCache[ip32]
    arpCacheLock.RUnlock()
    if !ok || time.Now().After(entry.validUntil) {
        return al.QueryIP(ip)
    }
    return entry.mac, nil
}

func queryIP(targetIP net.IP) (net.HardwareAddr, error) {
    var waiter chan bool
    
    //Check if already looking up
    ip32 := binary.BigEndian.Uint32(targetIP.To4())
    lookupCacheLock.Lock()
    for _, i := range lookupCache {
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
    //We are the first, thus we must also send the arp request
    if waiter == nil {
        waiter = make(chan bool)
        l := &lookup{ip: ip32, waiters: []chan bool{ waiter, }}
        lookupCache = append(lookupCache, l)
        pkt := sendArpQuery(targetIP)
        if pkt == nil {
            lookupCacheLock.Unlock()
            return nil, ErrInvalidAddressRequested
        }
    }
    lookupCacheLock.Unlock()
    
    //We block until either the channel is closed or we got a message.
    //The worker automatically puts the replies into the cache and notifies all channels who looked it up.
    _, _ = <-waiter
    
    return al.FindInCache(targetIP)
}

func sendArpQuery(targetIP net.IP) ethernet.Layer2Paket  {
    localIP := al.ip4.GetLocalAddressForSubnet(targetIP)
    if localIP == nil {
        return nil
    }
    p := make([]byte, 60)
    binary.BigEndian.PutUint16(p[0:2], 0x0001)
    binary.BigEndian.PutUint16(p[2:4], uint16(layers.EthernetTypeIPv4))
    p[4] = 6
    p[5] = 4
    binary.BigEndian.PutUint16(p[6:8], 0x0001)
    copy(p[8:14], srcMAC)
    copy(p[14:18], localIP)
    copy(p[18:24], BroadcastMACAddress)
    copy(p[24:28], targetIP)
    return p
}