package ipv4

import (
	"net"
    "sync"
	"github.com/arcpop/network/netdev"
	"github.com/arcpop/network/util"
)

type RoutingEntry struct {
    netmask uint32
    network uint32
    gateway uint32
    metric int
    flags int
    Iface netdev.Interface
}

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
    routingTable []*RoutingEntry
    routingTableLock sync.RWMutex
)

func RouteAddNet(from net.IPNet, gateway net.IP, metric, flags int, dev netdev.Interface) {
    ip32 := util.IPToUint32(from.IP)
    nm32 := util.IPToUint32(from.Mask)
    var gw32 uint32
    if gateway != nil {
        gw32 = util.IPToUint32(gateway)
        flags |= FlagGateway
    }
    e := &RoutingEntry{
        netmask: nm32,
        network: ip32,
        gateway: gw32,
        metric: metric,
        flags: flags,
        Iface: dev,
    }
    
    routingTableLock.Lock()
    routingTable = append(routingTable, e)
    routingTableLock.Unlock()
}

func RouteAddHost(host net.IP, gateway net.IP, metric, flags int, dev netdev.Interface) {
    RouteAddNet(net.IPNet{ IP: host, Mask: []byte{0xFF, 0xFF, 0xFF, 0xFF}}, gateway, metric, flags | FlagHost, dev)
}

func RouteDeleteInterface(iface netdev.Interface)  {
    routingTableLock.Lock()
    for k := 0; k < len(routingTable); k++ {
        v := routingTable[k]
        if v.Iface == iface {
            if k == len(routingTable) - 1 {
                routingTable = routingTable[:k]
            } else {
                routingTable = append(routingTable[:k], routingTable[k + 1:]...)
            }
        }
    }
    routingTableLock.Unlock()
}


func RoutingGetRoute(targetIP net.IP) (*RoutingEntry, error) {
    ip32 := util.IPToUint32(targetIP)
    routingTableLock.RLock()
    defer routingTableLock.RUnlock()
    
    var bestRoute *RoutingEntry
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
        return nil, ErrPacketNotRoutable
    }
    
    return bestRoute, nil
}

func ConfigureInterfaceAddress(ifname string, address net.IPNet) error {
    iface := netdev.InterfaceByName(ifname)
    if iface == nil {
        return ErrInterfaceNotFound
    }
    
    currentIP := iface.GetIPv4Address()
    if currentIP.IsGlobalUnicast() {
        RouteDeleteInterface(iface)
    }
    
    iface.SetIPv4Address(address.IP, net.IP(address.Mask))
    RouteAddHost(address.IP, nil, 0, 0, iface)
    RouteAddNet(address, nil, 1, 0, iface)
    return nil
}

func initRoutingTable() {
    
}