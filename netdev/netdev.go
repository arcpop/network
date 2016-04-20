package netdev

import (
	"net"
	"sync"
	"errors"
	"strconv"
	"github.com/arcpop/network/util"
	"bytes"
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

func GetInterfaceInfo(iface Interface) string {
    if iface == nil {
        return ""
    }
    str := "Interface " + iface.GetName() + "\n"
    hwAddr := iface.GetHardwareAddress()
    if hwAddr != nil {
        str += "\tHardware Address: " + hwAddr.String() + "\n"
    }
    ipv4 := iface.GetIPv4Address()
    nm := iface.GetIPv4Netmask()
    if ipv4 != nil && util.IPToUint32(ipv4) != 0 {
        str += "\tIPv4 Address: " + (&net.IPNet{ IP: ipv4, Mask: net.IPMask(nm)}).String() + "\n"
    }
    ipv6 := iface.GetIPv6Address()
    if ipv6 != nil && bytes.Compare(ipv6, make([]byte, 16)) != 0 {
        str += "\tIPv6 Address: " + ipv6.String() + "/" + strconv.Itoa(iface.GetIPv6Netmask()) + "\n"
    }
    str += "\tMTU: " + strconv.Itoa(iface.GetMTU()) + "\n"
    p, b, e := iface.GetTxStats()
    str += "\tTxPackets: " + strconv.FormatUint(p, 10) + " TxBytes: " + strconv.FormatUint(b, 10) + 
        " TxErrors: " + strconv.FormatUint(e, 10) + "\n"
    b, p, e = iface.GetRxStats()
    str += "\tRxPackets: " + strconv.FormatUint(p, 10) + " RxBytes: " + strconv.FormatUint(b, 10) + 
        " RxErrors: " + strconv.FormatUint(e, 10) + "\n"
    return str
}

func GetAllInterfaceInfo() string {
    str := ""
    interfaceListLock.RLock()
    for _, i := range interfaceList {
        str += GetInterfaceInfo(i) + "\n"
    }
    interfaceListLock.RUnlock()
    return str
}