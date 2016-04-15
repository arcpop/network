package arp

import (
	"net"
	"time"
	"sync"
    "github.com/arcpop/network/ethernet"
    "github.com/arcpop/network/ipv4"
	"log"
)


type arpCacheEntry struct {
    ip net.IP
    mac net.HardwareAddr
    validUntil time.Time
}

type ArpLayer struct {
    
    runningLock sync.RWMutex
    running bool
    ticker *time.Ticker
    el *ethernet.EthernetLayer
    ip4 *ipv4.IPv4Layer
    
    //arpCache is the arpcache to lookup mac addresses for given ip addresses
    arpCache map[uint32] *arpCacheEntry
    arpCacheLock sync.RWMutex
    
    //lookupCache contains all sent but not yet received arp requests
    lookupCache []*lookup
    lookupCacheLock sync.Mutex
    
    senderQueue chan *ethernet.Layer3Paket
}


func NewArpLayer() *ArpLayer {
    return &ArpLayer{ arpCache: make(map[uint32] *arpCacheEntry), lookupCache: make([]*lookup, 0, 10) }
}

func (al *ArpLayer) Start(el *ethernet.EthernetLayer)  {
    al.runningLock.Lock()
    defer al.runningLock.Unlock()
    if al.running {
        log.Println("Arp: Already running!")
        return
    }
    al.running = true
    al.el = el
    al.ticker = time.NewTicker(500 * time.Millisecond)
    go al.invalidateCache()
    al.runningLock.Unlock()
}


func (al *ArpLayer) invalidateCache()  {
    for _ = range al.ticker.C {
        al.arpCacheLock.Lock()
        for k, v := range al.arpCache {
            if time.Now().After(v.validUntil) {
                delete(al.arpCache, k)
                go al.Lookup(v.ip)
            }
        }
        al.arpCacheLock.Unlock()
    }
}