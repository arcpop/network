package arp

import (
    "net"
	"encoding/binary"
	"sync"
	"github.com/arcpop/network/ethernet"
	"github.com/arcpop/network/netdev"
	"errors"
	"time"
)

var (
    //BroadcastMACAddress is the broadcast hw address to send arp requests to.
    BroadcastMACAddress = net.HardwareAddr{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
)

const (
    DefaultTTL = 60,
    Timeout = 5,
)

const (
    waiting = iota
    resolved = iota
)

type arpCacheEntry struct {
    dev *netdev.NetDev
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

func SetIPAndSend(dev *netdev.NetDev, pkt []byte, targetIP net.IP) {
    arpCacheLock.Lock()
    e, ok := arpCache[targetIP.ToUint32()]
    if !ok {
        e = &arpCacheEntry{
            dev: dev,
            state: waiting,
            ttl: Timeout,
            retries: 5,
            queuedPackets: make(chan []byte, 1024),
        }
        e.queuedPackets <- pkt
        arpCache[targetIP.ToUint32()] = e
        arpCacheLock.Unlock()
        go arpRequest(targetIP, dev)
        return
    }
    e.queuedPackets <- pkt
    arpCacheLock.Unlock()
}
func arpCacheInsert(dev *netdev.NetDev, ip net.IP, mac net.HardwareAddr)  {
    ae := & arpCacheEntry{
        state: resolved,
        mac: make([]byte, 6),
        dev: dev,
        ttl: DefaultTTL,
    }
    copy(ae.mac, mac)
    arpCacheLock.Lock()
    oldEntry, ok := arpCache[ToUint32(ip)]
    arpCache[ToUint32(ip)] = ae
    arpCacheLock.Unlock()
    if ok {
        go sendQueuedPackets(oldEntry)
    }
}

func sendQueuedPackets(e *arpCacheEntry) {
    if e.queuedPackets != nil {
        for {
            select {
                case pkt := <- e.queuedPackets:
                    copy(pkt[ethernet.HeaderLength + 12:], net.IP)
                    e.dev.TxPacket(e.dev, pkt)
                default:
                    close(e.queuedPackets)
                    return
            }
        }
    }
}

func dropQueuedPackets(e *arpCacheEntry) {
    dropped = 0
    if e.queuedPackets != nil && len(e.queuedPackets) > 0 {
        for {
            select {
                case pkt := <- e.queuedPackets:
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
                v.retry--
                if (v.state == waiting && v.retry < 0) || (v.state == resolved) {
                    if v.state == waiting {
                        dropQueuedPackets(v)
                    }
                    delete(arpCache, k)
                } else {
                    v.ttl = Timeout
                    arpRequest(ToIP(k), v.dev)
                }
            }
        }
        arpCacheLock.Unlock()
    }
}

func (ip net.IP) ToUint32() uint32 {
    return binary.BigEndian.Uint32(ip)
}

func ToIP(u uint32) net.IP {
    var buf [4]byte
    binary.BigEndian.PutUint32(buf[:], u)
    return net.IP(buf)
}