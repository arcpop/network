package netdev

import (
	"net"
	"sync"
	"errors"
)

type Interface interface {
    RxPacket() []byte
    TxPacket(pkt []byte)
    
    GetName()string
    GetMTU() int
    
    GetTxStats() (pkts uint64, bytes uint64, errors uint64)
    GetRxStats() (pkts uint64, bytes uint64, errors uint64)
    
    GetIPv4Address() net.IP
    GetIPv4Netmask() net.IP
    SetIPv4Address(ip, netmask net.IP)
    
    GetIPv6Address() net.IP
    GetIPv6Netmask() int
    SetIPv6Address(ip net.IP, netmask int)
    
    GetHardwareAddress() net.HardwareAddr
    
    Close()
}

type InterfaceStats struct {
    TxPackets, TxBytes, TxErrors uint64
    RxPackets, RxBytes, RxErrors uint64
}

var ErrLoopbackAlreadyExists = errors.New("Netdev: Loopback device already exists!")

func NewLoopback(name string) (Interface, error) {
    l := &loopback { name: name, }
    
    interfaceListLock.Lock()
    defer interfaceListLock.Unlock()
    
    if loopbackExists {
        return nil, ErrLoopbackAlreadyExists
    }
    
    loopbackExists = true
    interfaceList = append(interfaceList, l)
    return l, nil
}

var interfaceListLock sync.RWMutex
var interfaceList []Interface
var loopbackExists = false


func ShutdownInterfaces()  {
    interfaceListLock.Lock()
    loopbackExists = false
    for _, v := range interfaceList {
        v.Close()
    }
    interfaceList = nil
    interfaceListLock.Unlock()
}

func InterfaceByName(name string) Interface {
    interfaceListLock.RLock()
    defer interfaceListLock.RUnlock()
    for _, v := range interfaceList {
        if v.GetName() == name {
            return v
        }
    }
    return nil
}