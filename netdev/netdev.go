package netdev

import (
	"net"
	"github.com/arcpop/network/tap"
	"github.com/arcpop/network/config"
	"sync/atomic"
)


type NetDev struct {
    RxPacket func (dev *NetDev) []byte
    TxPacket func (dev *NetDev, pkt []byte)
    
    GetName func () string
    GetMTU func() int
    
    GetIPv4Address func () net.IP
    GetIPv4Netmask func () net.IP
    SetIPv4Address func (ip, netmask net.IP)
    
    GetIPv6Address func () net.IP
    GetIPv6Netmask func () int
    SetIPv6Address func (ip net.IP, netmask int)
    
    GetHardwareAddress func () net.HardwareAddr
    
    Stats struct {
        TxPackets, TxBytes, TxErrors uint64
        RxPackets, RxBytes, RxErrors uint64
    }
    vether *VEthernet
}

type VEthernet struct {
    txQueue chan []byte
    rxQueue chan []byte
    tapDev *tap.Adapter
    name string
    mtu int
    macAddr net.HardwareAddr
    ipv4addr net.IP
    ipv4nm net.IP
    ipv6addr net.IP
    ipv6nm int
}

func NewLoopbackDevice() (*NetDev, error) {
    return &NetDev {
        RxPacket: loopbackRxPacket,
        TxPacket: loopbackTxPacket,
        GetName: func () string { return "lo" },
        GetMTU: func () int { return 1500 },
        GetIPv4Address: func() net.IP { return net.IP{127, 0, 0, 1} },
        GetIPv4Netmask: func() net.IP { return net.IP{255, 0, 0, 0} },
        SetIPv4Address: func(ip, netmask net.IP) { return },
        GetIPv6Address: func () net.IP { return net.IP{0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1} },
        GetIPv6Netmask: func () int { return 128 },
        SetIPv6Address: func (ip net.IP, netmask int) { return },
        GetHardwareAddress: func () net.HardwareAddr { return nil },
    }, nil
}

func NewTapDevice(name string) (*NetDev, error) {
    adapter, err := tap.NewAdapter(name)
    if err != nil {
        return nil, err
    }
    vether := &VEthernet{
        txQueue: make(chan []byte, config.Ethernet.TxQueueSize),
        rxQueue: make(chan []byte, config.Ethernet.RxQueueSize),
        tapDev: adapter,
        name: name,
        mtu: adapter.GetMTU(),
        macAddr: adapter.GetHWAddr(),
    }
    adapter.StartProcessing(vether.txQueue, vether.rxQueue)
    netDev := &NetDev{
        vether: vether,
        RxPacket: vetherRxPacket,
        TxPacket: vetherTxPacket,
        GetName: func () string { return vether.name },
        GetMTU: func () int { return vether.mtu },
        GetIPv4Address: func() net.IP { return vether.ipv4addr },
        GetIPv4Netmask: func() net.IP { return vether.ipv4nm },
        SetIPv4Address: func(ip, netmask net.IP) { 
            vether.ipv4addr = ip
            vether.ipv4nm = netmask
        },
        GetIPv6Address: func () net.IP { return vether.ipv6addr },
        GetIPv6Netmask: func () int { return vether.ipv6nm },
        SetIPv6Address: func (ip net.IP, netmask int) { 
            vether.ipv6addr = ip
            vether.ipv6nm = netmask
        },
        GetHardwareAddress: func () net.HardwareAddr { return vether.macAddr },
    }
    return netDev, nil
}



func vetherRxPacket(dev *NetDev) []byte {
    pkt := <- dev.vether.rxQueue
    atomic.AddUint64(&dev.Stats.RxBytes, uint64(len(pkt)))
    atomic.AddUint64(&dev.Stats.RxPackets, 1)
    return pkt
}

func vetherTxPacket(dev *NetDev, pkt []byte) {
    l := uint64(len(pkt))
    select {
        case dev.vether.txQueue <- pkt:
            atomic.AddUint64(&dev.Stats.TxBytes, l)
            atomic.AddUint64(&dev.Stats.TxPackets, 1)
        default:
            atomic.AddUint64(&dev.Stats.TxErrors, 1)
    }
}