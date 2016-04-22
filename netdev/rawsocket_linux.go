// +build linux

package netdev

import (
	"net"
    "syscall"
	"log"
	"sync/atomic"
	"sync"
	"github.com/arcpop/network/config"
	"errors"
	"unsafe"
)

type rawsock struct {
    fd int
    iface *net.Interface
    
    ipv4Lock sync.RWMutex
    ipv4 net.IP
    netmaskv4 net.IP
    
    ipv6Lock sync.RWMutex
    ipv6 net.IP
    netmaskv6 int
    
    TxPackets, TxBytes, TxErrors uint64
    RxPackets, RxBytes, RxErrors uint64
    
    RxQueue chan []byte
    TxQueue chan []byte
}

var ErrDeviceAlreadyExists = errors.New("Rawsocket: A rawsocket implementation for that device already exists!")

type ifrfl struct {
    ifrname [syscall.IFNAMSIZ]byte
    ifrflags int16
}
func ioctl(fd, code int, p uintptr) error {
    _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(code), p)
    if errno == 0 {
        return nil
    }
    return errno
}

func NewRawSocket(ifname string) (Interface, error) {
    interfaceListLock.Lock()
    defer interfaceListLock.Unlock()
    for _, v := range interfaceList {
        if v.GetName() == ifname {
            return nil, ErrDeviceAlreadyExists
        }
    }
    
    //htons(ETH_P_ALL) = htons(0x0003) = 0x0300
    iface, err := net.InterfaceByName(ifname)
    if err != nil {
        return nil, err
    }
    
    sockaddrll := &syscall.SockaddrLinklayer {
        Protocol: 0x0300,
        Ifindex: iface.Index,
    }
    
    fd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, 0x0300) 
    if err != nil {
        return nil, err
    }
    
    err = syscall.Bind(fd, sockaddrll)
    if err != nil {
        syscall.Close(fd)
        return nil, err
    }
    
    ifr := ifrfl{}
    copy(ifr.ifrname[:], []byte(ifname))
    ifr.ifrname[syscall.IFNAMSIZ - 1] = 0
    err = ioctl(fd, syscall.SIOCGIFFLAGS, uintptr(unsafe.Pointer(&ifr)))
    if err != nil {
        syscall.Close(fd)
        return nil, err
    }
    
    ifr.ifrflags |= syscall.IFF_UP
    
    copy(ifr.ifrname[:], []byte(ifname))
    ifr.ifrname[syscall.IFNAMSIZ - 1] = 0
    err = ioctl(fd, syscall.SIOCSIFFLAGS, uintptr(unsafe.Pointer(&ifr)))
    if err != nil {
        syscall.Close(fd)
        return nil, err
    }
    
    
    rs := &rawsock{ 
        fd: fd, 
        iface: iface, 
        RxQueue: make(chan []byte, config.Device.RxQueueSize), 
        TxQueue: make(chan []byte, config.Device.TxQueueSize),
    }
    
    for i := 0; i < config.Device.RxQueueWorkers; i++ {
        go rs.rxPacketWorker()
    }
    for i := 0; i < config.Device.TxQueueWorkers; i++ {
        go rs.txPacketWorker()
    }
    
    interfaceList = append(interfaceList, rs)
    return rs, nil
}

func (rs *rawsock) rxPacketWorker() {
    log.Println("RawSocket: RxWorker starting!")
    for {
        //MTU + ethernet header size
        pkt := make([]byte, rs.iface.MTU + 14)
        n, _, err := syscall.Recvfrom(rs.fd, pkt, 0)
        if err != nil {
            log.Println("RawSocket.Recvfrom: ", err)
            atomic.AddUint64(&rs.RxErrors, 1)
            continue
        }
        atomic.AddUint64(&rs.RxPackets, 1)
        atomic.AddUint64(&rs.RxBytes, uint64(n))
        rs.RxQueue <- pkt[:n]
    }
}
func (rs *rawsock) txPacketWorker() {
    log.Println("RawSocket: TxWorker starting!")
    sockaddrll := &syscall.SockaddrLinklayer{
        Ifindex: rs.iface.Index,
        Protocol: 0x0300,
        Halen: 6,
    }
    for pkt := range rs.TxQueue {
        copy(sockaddrll.Addr[0:6], pkt[0:6])
        err := syscall.Sendto(rs.fd, pkt, 0, sockaddrll)
        if err != nil {
            log.Println("RawSocket.Sendto: ", err)
            atomic.AddUint64(&rs.TxErrors, 1)
            continue
        }
        atomic.AddUint64(&rs.TxPackets, 1)
        atomic.AddUint64(&rs.TxBytes, uint64(len(pkt)))
    }
}

func (rs *rawsock) RxPacket() []byte {
    return <- rs.RxQueue
}
func (rs *rawsock) TxPacket(pkt []byte) {
    rs.TxQueue <- pkt
}

func (rs *rawsock) GetName()string {
    return rs.iface.Name
}

func (rs *rawsock) GetMTU() int {
    return rs.iface.MTU
}

func (rs *rawsock) GetTxStats() (pkts uint64, bytes uint64, errors uint64) {
    return rs.TxPackets, rs.TxBytes, rs.TxErrors
}

func (rs *rawsock) GetRxStats() (pkts uint64, bytes uint64, errors uint64) {
    return rs.RxPackets, rs.RxBytes, rs.RxErrors
}

func (rs *rawsock) GetIPv4Address() net.IP {
    var ip [4]byte
    rs.ipv4Lock.RLock()
    copy(ip[:], rs.ipv4[:])
    rs.ipv4Lock.RUnlock()
    return net.IP(ip[:])
}

func (rs *rawsock) GetIPv4Netmask() net.IP {
    var nm [4]byte
    rs.ipv4Lock.RLock()
    copy(nm[:], rs.netmaskv4[:])
    rs.ipv4Lock.RUnlock()
    return net.IP(nm[:])
}

func (rs *rawsock) SetIPv4Address(ip, netmask net.IP) {
    rs.ipv4Lock.Lock()
    rs.ipv4 = make([]byte, 4)
    copy(rs.ipv4[:], ip.To4())
    rs.netmaskv4 = make([]byte, 4)
    copy(rs.netmaskv4[:], netmask.To4())
    rs.ipv4Lock.Unlock()
}

func (rs *rawsock) GetIPv6Address() net.IP {
    var ip [16]byte
    rs.ipv6Lock.RLock()
    copy(ip[:], rs.ipv6[:])
    rs.ipv6Lock.RUnlock()
    return net.IP(ip[:])
}

func (rs *rawsock) GetIPv6Netmask() int {
    return rs.netmaskv6
}

func (rs *rawsock) SetIPv6Address(ip net.IP, netmask int) {
    rs.ipv6Lock.Lock()
    rs.ipv6 = make([]byte, 16)
    copy(rs.ipv6, ip)
    rs.netmaskv6 = netmask
    rs.ipv6Lock.Unlock()
}

func (rs *rawsock) GetHardwareAddress() net.HardwareAddr {
    var mac [6]byte
    copy(mac[:], rs.iface.HardwareAddr)
    return net.HardwareAddr(mac[:])
}


func (rs *rawsock) Close() {
    close(rs.TxQueue)
    close(rs.RxQueue)
    syscall.Close(rs.fd)
    return
}