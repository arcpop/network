package netdev

import (
	"sync/atomic"
    "net"
)

type loopback struct {
    name string
    TxPackets, TxBytes, TxErrors uint64
    RxPackets, RxBytes, RxErrors uint64
}

func (l *loopback)GetName() string { 
    return l.name
}

func (*loopback)GetMTU() int { 
    return 65535 
}

func (*loopback)GetIPv4Address() net.IP { 
    return net.IP{127, 0, 0, 1} 
}

func (*loopback)GetIPv4Netmask() net.IP { 
    return net.IP{255, 0, 0, 0} 
}

func (*loopback)SetIPv4Address(ip, netmask net.IP) { 
    return 
}

func (*loopback)GetIPv6Address() net.IP { 
    return net.IP{0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1} 
}

func (*loopback)GetIPv6Netmask() int { 
    return 128 
}

func (*loopback)SetIPv6Address(ip net.IP, netmask int) { 
    return 
}

func (*loopback)GetHardwareAddress() net.HardwareAddr { 
    return nil 
}
func (l *loopback)GetTxStats() (pkts uint64, bytes uint64, errors uint64) {
    return l.TxPackets, l.TxBytes, l.TxErrors
}
func (l *loopback)GetRxStats() (pkts uint64, bytes uint64, errors uint64) {
    return l.RxPackets, l.RxBytes, l.RxErrors
}

var loopbackQueue = make(chan []byte, 1024)

func (l *loopback) RxPacket() []byte {
    pkt := <- loopbackQueue
    atomic.AddUint64(&l.RxBytes, uint64(len(pkt)))
    atomic.AddUint64(&l.RxPackets, 1)
    return pkt
}

func (l *loopback) TxPacket(pkt []byte) {
    s := uint64(len(pkt))
    select {
        case loopbackQueue <- pkt:
            atomic.AddUint64(&l.TxBytes, s)
            atomic.AddUint64(&l.TxPackets, 1)
        default:
            atomic.AddUint64(&l.TxErrors, 1)
    }
}

func (l* loopback) Close()  {
    return
}