package ipv4

import (
	"net"
    "sync"
	"github.com/arcpop/network/netdev"
	"github.com/arcpop/network/util"
	"github.com/arcpop/network/arp"
)

const (
    MetricLocalhost = 0
    MetricMin = 1
    MetricDefault = 1024
    MetricMax = 1 << 20
    metricOverMax = MetricMax + 1
)
const (
    FlagHost = 1 << iota
    FlagGateway = 1 << iota
)
var (
    ErrInterfaceNotFound = errors.New("Interface not found")
)

func RouteAddNet(from net.IPNet, gateway net.IP, metric, flags int, ifname string) error {
    dev := netdev.NetdevByName(ifname)
    if dev == nil {
        return ErrInterfaceNotFound
    }
    ip32 := util.IPToUint32(from.IP)
    nm32 := util.IPToUint32(from.Mask)
    gw32 = 0
    if gateway != nil {
        gw32 = util.IPToUint32(gateway)
        flags |= FlagGateway
    }
    e := &routingEntry{
        netmask: nm32,
        network: ip32,
        gateway: gw32,
        metric: metric,
        flags: flags,
        routeDev: dev,
    }
    
    routingTableLock.Lock()
    routingTable = append(routingTable, e)
    routingTableLock.Unlock()
}

func RouteAddHost(host net.IP, gateway net.IP, metric, flags int, ifname string) error {
    return RouteAddNet(net.IPNet{ IP: host, Mask: []byte{0xFF, 0xFF, 0xFF, 0xFF}}, gateway, metric, flags | FlagHost, ifname)
}

type routingEntry struct {
    netmask uint32
    network uint32
    gateway uint32
    metric int
    flags int
    routeDev* netdev.NetDev
}

var (
    routingTable []*routingEntry
    routingTableLock sync.RWMutex
)

func routingSendPacket(pkt []byte, ipHeader *Header) (error) {
    ip32 := util.IPToUint32(ipHeader.TargetIP)
    routingTableLock.RLock()
    defer routingTableLock.RUnlock()
    
    var bestRoute *routingEntry
    bestMetric := metricOverMax
    for _, e := range routingTable {
        if (e.network & e.netmask) == (ip32 & e.netmask) {
            if bestMetric > e.metric {
                bestRoute = e
                bestMetric = e.metric
            }
        }
    }
    if bestMetric == metricOverMax {
        return ErrPacketNotRoutable
    }
    if (bestRoute.flags & FlagGateway) != 0 {
        arp.SetIPAndSend(bestRoute.routeDev, pkt, bestRoute.gateway)
    } else {
        arp.SetIPAndSend(bestRoute.routeDev, pkt, ipHeader.TargetIP)
    }
}


func initRoutingTable() {
    
}