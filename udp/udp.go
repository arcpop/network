package udp

import (
    "github.com/arcpop/network/conn"
	"net"
    "sync"
	"math/rand"
	"errors"
	"github.com/arcpop/network/config"
	"github.com/arcpop/network/util"
	"github.com/arcpop/network/ipv4"
	"github.com/arcpop/network/ipv6"
)

type udpConnection struct {
    lport, rport uint16
    identification uint16
    recvQueue chan []byte
    recvPartialPacket []byte
    readLock, writeLock sync.Mutex
    remoteIP, localIP net.IP
    isIPv4 bool
}

var (
    ErrLocalPortAlreadyBound = errors.New("Local port is already bound!")
    ErrInvalidPort = errors.New("Invalid remote port!")
    ErrNoNameResolution = errors.New("No name resolution installed!")
)

var (
    udpConnections4 map[uint16]*udpConnection
    udpConnections4Lock sync.RWMutex
    udpConnections6 map[uint16]*udpConnection
    udpConnections6Lock sync.RWMutex
    
    udpRecvQueue4 chan *ipv4.L3Packet
)

func Start()  {
    udpConnections4 = make(map[uint16]*udpConnection)
    udpConnections6 = make(map[uint16]*udpConnection)
    udpRecvQueue4 = make(chan *ipv4.L3Packet, config.UDP.RecvQueueSize)
}
func DialUDP(remoteAddr, localAddr string) (conn.Conn, error) {
    return nil, ErrNoNameResolution
}

func CreateUDP4(remoteIP net.IP, remotePort, localPort uint16) (conn.Conn, error)  {
    if remotePort == 0 {
        return nil, ErrInvalidPort
    }
    ip4 := remoteIP.To4()
    if ip4 == nil {
        return nil, ipv6.ErrNotImplemented
    }
    route, err := ipv4.RoutingGetRoute(ip4)
    if err != nil {
        return nil, err
    }
    if localPort == 0 {
        udpConnections4Lock.Lock()
        localPort := uint16(rand.Uint32() & 0xFFFF);
        _, ok := udpConnections4[localPort]
        for ok {
            localPort = uint16(rand.Uint32() & 0xFFFF);
            _, ok = udpConnections4[localPort]
        }
    } else {
        udpConnections4Lock.Lock()
        _, ok := udpConnections4[localPort]
        if ok {
            udpConnections4Lock.Unlock()
            return nil, ErrLocalPortAlreadyBound
        }
    }
    connection := &udpConnection{
        rport: remotePort,
        lport: localPort,
        recvQueue: make(chan []byte, config.UDP.ConnectionRecvQueueSize),
        remoteIP: make([]byte, 4),
        localIP: make([]byte, 4),
        isIPv4: true,
    }
    copy(connection.remoteIP, ip4)
    copy(connection.localIP, route.Iface.GetIPv4Address())
    return connection, nil
}

func (u *udpConnection) Read(b []byte) (n int, err error) {
    u.readLock.Lock()
    defer u.readLock.Unlock()
    empty, n := util.Drain(u.recvPartialPacket, b)
    if empty {
        u.recvPartialPacket = nil
        b = b[n:]
    } else {
        u.recvPartialPacket = u.recvPartialPacket[n:]
        return n, nil
    }
    i := n
    for {
        pkt := <- u.recvQueue
        empty, n := util.Drain(u.recvPartialPacket, b)
        if empty {
            u.recvPartialPacket = nil
            b = b[n:]
            i += n
        } else {
            u.recvPartialPacket = pkt[n:]
            return i, nil
        }
    }
}

func (u *udpConnection) Write(b []byte) (n int, err error) {
    u.writeLock.Lock()
    defer u.writeLock.Unlock()
    if u.isIPv4 {
        pkt := ipv4.AllocatePacket(len(b))
        copy(pkt.ProtocolData, b[:])
        header := &ipv4.Header{
            TargetIP: u.remoteIP,
            SourceIP: u.localIP,
            Identification: u.identification,
        }
        u.identification++
        pkt.IPHeader = header
        ipv4.Send(pkt)
        return len(b), nil
    } 
    return 0, ipv6.ErrNotImplemented
}


func udpRecvWorker()  {
    
}