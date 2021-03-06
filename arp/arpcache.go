package arp

import (
    "net"
	"sync"
	"github.com/arcpop/network/netdev"
	"bytes"
	"time"
	"log"
	"github.com/arcpop/network/util"
	"encoding/binary"
)

var (
    //BroadcastMACAddress is the broadcast hw address to send arp requests to.
    BroadcastMACAddress = net.HardwareAddr{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
)

const (
    DefaultTTL = 3600
    Timeout = 5
)

const (
    waiting = iota
    resolved = iota
)

type arpCacheEntry struct {
    dev netdev.Interface
    state int
    mac net.HardwareAddr
    ttl int
    retries int
    queuedPackets chan []byte
}

var (
    arpCache map[uint32] *arpCacheEntry
    arpCacheLock sync.RWMutex
)

//SetMACAndSend should be used by ipv4 layer to send packets. They get the destination mac assigned automatically.
func SetMACAndSend(dev netdev.Interface, pkt []byte, targetIP net.IP) {
    arpCacheLock.Lock()
    ip32 := util.IPToUint32(targetIP)
    e, ok := arpCache[ip32]
    if !ok {
        e = &arpCacheEntry{
            dev: dev,
            state: waiting,
            ttl: Timeout,
            retries: 5,
            queuedPackets: make(chan []byte, 1024),
        }
        e.queuedPackets <- pkt
        arpCache[ip32] = e
        arpCacheLock.Unlock()
        go arpRequest(targetIP, dev)
        return
    }
    arpCacheLock.Unlock()
    copy(pkt[0:6], e.mac)
    copy(pkt[6:12], e.dev.GetHardwareAddress())
    binary.BigEndian.PutUint16(pkt[12:14], 0x0800)
    e.dev.TxPacket(pkt)
}
func arpCacheInsert(dev netdev.Interface, ip net.IP, mac net.HardwareAddr)  {
    ip32 := util.IPToUint32(ip)
    ae := & arpCacheEntry{
        state: resolved,
        mac: make([]byte, 6),
        dev: dev,
        ttl: DefaultTTL,
    }
    copy(ae.mac, mac)
    arpCacheLock.Lock()
    oldEntry, ok := arpCache[ip32]
    arpCache[ip32] = ae
    arpCacheLock.Unlock()
    if ok {
        go sendQueuedPackets(oldEntry)
    }
}
func PassiveLearn(iface netdev.Interface, ip net.IP, mac net.HardwareAddr)  {
    ip32 := util.IPToUint32(ip)
    arpCacheLock.Lock()
    e, ok := arpCache[ip32]
    if !ok {
        arpCacheLock.Unlock()
        return
    }
    if e.state == waiting {
        arpCacheLock.Unlock()
        go arpCacheInsert(iface, ip, mac)
        return
    }
    //Check if there was some left over wrong entry
    if bytes.Compare(mac, e.mac) != 0 {
        delete(arpCache, ip32)
        arpCacheLock.Unlock()
        go arpCacheInsert(iface, ip, mac)
        return
    }
    e.ttl = DefaultTTL
    arpCacheLock.Unlock()
}

func cacheUpdate(dev netdev.Interface, ip net.IP, mac net.HardwareAddr) {
    ip32 := util.IPToUint32(ip)
    arpCacheLock.Lock()
    e, ok := arpCache[ip32]
    if !ok || e.state == waiting {
        arpCacheLock.Unlock()
        arpCacheInsert(dev, ip, mac)
        return
    }
    //Check if there was some left over wrong entry
    if bytes.Compare(mac, e.mac) != 0 {
        delete(arpCache, ip32)
        arpCacheLock.Unlock()
        arpCacheInsert(dev, ip, mac)
        return
    }
    e.ttl = DefaultTTL
    arpCacheLock.Unlock()
}

func sendQueuedPackets(e *arpCacheEntry) {
    if e.queuedPackets != nil {
        for {
            select {
                case pkt := <- e.queuedPackets:
                    copy(pkt[0:6], e.mac)
                    copy(pkt[6:12], e.dev.GetHardwareAddress())
                    binary.BigEndian.PutUint16(pkt[12:14], 0x0800)
                    e.dev.TxPacket(pkt)
                default:
                    close(e.queuedPackets)
                    return
            }
        }
    }
}

func dropQueuedPackets(e *arpCacheEntry) {
    dropped := 0
    if e.queuedPackets != nil && len(e.queuedPackets) > 0 {
        for {
            select {
                case _ = <- e.queuedPackets:
                    dropped++
                default:
                    close(e.queuedPackets)
                    log.Println("Arp: Dropped ", dropped, " packets due to not resolving arp request.")
                    return
            }
        }
    }
}

func arpTicker() {
    tckr := time.NewTicker(time.Second)
    for _ = range tckr.C {
        arpCacheLock.Lock()
        for k,v := range arpCache {
            v.ttl--
            if v.ttl <= 0 {
                v.retries--
                if (v.state == waiting && v.retries < 0) || (v.state == resolved) {
                    if v.state == waiting {
                        dropQueuedPackets(v)
                    }
                    delete(arpCache, k)
                } else {
                    v.ttl = Timeout
                    arpRequest(util.ToIP(k), v.dev)
                }
            }
        }
        arpCacheLock.Unlock()
    }
}

func GetCacheAsString() string {
    res := "IP - MAC\n"
    arpCacheLock.RLock()
    defer arpCacheLock.RUnlock()
    for i, e := range arpCache {
        res += util.ToIP(i).String() + " - " + e.mac.String() + "\n"
    }
    return res
}

func QueryIP(ip, dev string, retries int) bool {
    ipaddr := net.ParseIP(ip).To4()
    iface := netdev.InterfaceByName(dev)
    if iface == nil || ipaddr == nil {
        return false
    }
    ip32 := util.IPToUint32(ipaddr)
    arpCacheLock.Lock()
    _, ok := arpCache[ip32]
    if ok {
        arpCacheLock.Unlock()
        return true
    }
    e := &arpCacheEntry{
        dev: iface,
        state: waiting,
        ttl: Timeout,
        retries: retries,
        queuedPackets: make(chan []byte, 1024),
    }
    arpCache[ip32] = e
    arpCacheLock.Unlock()
    arpRequest(ipaddr, iface)
    return true
}
