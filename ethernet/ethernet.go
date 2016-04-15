//Package ethernet represents the network stack at layer 2
package ethernet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/arcpop/network/tap"
	"github.com/arcpop/network/util"
	"github.com/google/gopacket/layers"
	"log"
	"net"
	"sync"
)

//Layer3Paket represents an OSI layer 3 packet. Since we do not want ARP stuff in here to support 
//also IPv6 and maybe others, the upper layer has to supply us with a destination mac address.
type Layer3Paket struct {
	Data []byte
	DstMAC net.HardwareAddr
}

type protocol struct {
	sendQueue chan *Layer3Paket
	recvQueue chan []byte
	numberOfSendQueues int
}

//EthernetLayer represents the control structure at ethernet layer (OSI layer 2)
type EthernetLayer struct {
	NumberOfQueues int
	sendQueueSize, receiveQueueSize int
	
	processingLock sync.RWMutex
	processing bool
	ta *tap.Adapter
	mtu            int
	stop chan bool
	receiveQueue chan []byte
	sendQueue    chan []byte

	addressLock sync.RWMutex
	address     net.HardwareAddr

	supportedProtocolQueuesLock sync.RWMutex
	supportedProtocolQueues map[uint16]*protocol
}

//NewEthernetLayer creates a new EthernetLayer
func NewEthernetLayer(numberOfQueues, MTU, sendQueueSize, receiveQueueSize int, hwAddr net.HardwareAddr) (el *EthernetLayer) {
	el = &EthernetLayer{
		NumberOfQueues:   numberOfQueues,
		mtu:              MTU,
		sendQueueSize:    sendQueueSize,
		receiveQueueSize: receiveQueueSize,
		address: make([]byte, 6),
	}
	copy(el.address, hwAddr)
	return el
}

//RegisterProtocolQueues registers the send and recv queues for a given ethernet type protocol.
func (el *EthernetLayer) RegisterProtocolQueues(protocolType layers.EthernetType, numberOfSendQueues int, protocolSendQueue chan *Layer3Paket, protocolRecvQueue chan []byte) {
	el.processingLock.RLock()
	el.supportedProtocolQueuesLock.Lock()
	_, ok := el.supportedProtocolQueues[uint16(protocolType)]
	if ok {
		log.Println("Protocol was already registered, overwriting queues!")
	}
	el.supportedProtocolQueues[uint16(protocolType)] = &protocol{
		sendQueue: protocolSendQueue, 
		recvQueue: protocolRecvQueue, 
		numberOfSendQueues: numberOfSendQueues,
	}
	
	//If we haven't seen this protocol before and we are already processing, start the send processing queues
	if !ok {
		if el.processing {
			for i := 0; i < numberOfSendQueues; i++ {
				go el.sendEthernet(uint16(protocolType))
			}
		}
	}
	
	el.supportedProtocolQueuesLock.Unlock()
	el.processingLock.RUnlock()
}

//UnregisterProtocol unregisters an associated protocol and closes the coresponding channels if closeQueues is true
func (el *EthernetLayer) UnregisterProtocol(protocolType layers.EthernetType, closeQueues bool) {
	el.supportedProtocolQueuesLock.Lock()
	pq, ok := el.supportedProtocolQueues[uint16(protocolType)]
	if !ok {
		log.Println("Ethernet: Tried to unregister protocol which is not registered: ", protocolType)
	} else {
		delete(el.supportedProtocolQueues, uint16(protocolType))
		if closeQueues {
			close(pq.recvQueue)
			close(pq.sendQueue)
		}
	}
	el.supportedProtocolQueuesLock.Unlock()
}

//StartProcessing starts the processing queues of the EthernetLayer.
func (el *EthernetLayer) StartProcessing(ta *tap.Adapter) {
	el.processingLock.Lock()
	el.ta = ta
	el.mtu = ta.MTU
	el.sendQueue = make(chan []byte, el.sendQueueSize)
	el.receiveQueue = make(chan []byte, el.receiveQueueSize)
	ta.StartProcessing(el.sendQueue, el.receiveQueue)
	for i := 0; i < el.NumberOfQueues; i++ {
		go el.receiveEthernet()
	}
	for t,pq := range el.supportedProtocolQueues {
		for i := 0; i < pq.numberOfSendQueues; i++ {
			go el.sendEthernet(t)
		}
	}
	el.processing = true
	el.processingLock.Unlock()
}

//StopProcessing stops the processing of the EthernetLayer processing queues.
func (el *EthernetLayer) StopProcessing() {
	el.processingLock.Lock()
	el.supportedProtocolQueuesLock.Lock()
	if el.processing {
		close(el.sendQueue)
		close(el.receiveQueue)
		el.processing = false
		el.ta.StopProcessing()
		el.ta = nil
		el.mtu = 0
	}
	el.supportedProtocolQueuesLock.Unlock()
	el.processingLock.Unlock()
}

//GetMACAddress returns the associated MAC address of the interface we operate on.
func (el *EthernetLayer) GetMACAddress() (net.HardwareAddr) {
	var buf [6]byte
	el.addressLock.RLock()
	copy(buf[:], el.address)
	el.addressLock.RUnlock()
	return net.HardwareAddr(buf[:])
}

//Header represents an ethernet header
type Header struct {
	DstMAC       net.HardwareAddr
	SrcMAC       net.HardwareAddr
	EthernetType uint16
	dataOffset   int
}

//ErrInvalidPacket gets returned if the packet is too short
var ErrInvalidPacket = errors.New("Malformed ethernet packet (too short)")

func decodeFromBytes(p []byte, eth *Header) error {
	if len(p) < 14 {
		return ErrInvalidPacket
	}
	eth.DstMAC = net.HardwareAddr(p[0:6])
	eth.SrcMAC = net.HardwareAddr(p[6:12])
	eth.EthernetType = binary.BigEndian.Uint16(p[12:14])
	eth.dataOffset = 14
	return nil
}

func (el *EthernetLayer) receiveEthernet() {
	eth := &Header{}
	for {
		packet, ok := <-el.receiveQueue
		if !ok {
			return
		}

		err := decodeFromBytes(packet, eth)
		if err != nil {
			log.Println("Ethernet: Failed to decode a packet!", err)
		}
		el.addressLock.RLock()
		cmp := bytes.Compare(el.address, eth.DstMAC)
		el.addressLock.RUnlock()
		if cmp != 0 {
			log.Println("Ethernet: Got packet with wrong MAC-Address: ", eth.DstMAC.String())
			continue
		}
		el.supportedProtocolQueuesLock.RLock()
		pq, ok := el.supportedProtocolQueues[eth.EthernetType]
		if !ok {
			el.supportedProtocolQueuesLock.RUnlock()
			log.Println("Ethernet: No handler found for protocol: ", eth.EthernetType)
			continue
		}
		pq.recvQueue <- packet[14:]
		el.supportedProtocolQueuesLock.RUnlock()
	}
}

func (el *EthernetLayer) sendEthernet(protocolType uint16) {
	el.processingLock.RLock()
	var pkt = make([]byte, el.mtu)
	el.processingLock.RUnlock()
	for {
		if util.ChannelClosed(el.stop) {
			return
		}
		el.supportedProtocolQueuesLock.RLock()
		pq, ok := el.supportedProtocolQueues[protocolType]
		if !ok {
			//The protocol got unregistered, bye bye
			el.supportedProtocolQueuesLock.RUnlock()
			return
		}
		l3pkt := <-pq.sendQueue
		totalLength := len(l3pkt.Data) + 14
		if totalLength > el.mtu {
			log.Println("Ethernet: Packet with an invalid size received. Paket will be dropped.")
			el.supportedProtocolQueuesLock.RUnlock()	
			continue
		}
		
		copy(pkt[0:6], l3pkt.DstMAC)
		el.addressLock.RLock()
		copy(pkt[6:12], el.address)
		el.addressLock.RUnlock()
		binary.BigEndian.PutUint16(pkt[12:14], protocolType)
		copy(pkt[14:], l3pkt.Data)
		el.sendQueue <- pkt
		pkt = make([]byte, el.mtu)
		el.supportedProtocolQueuesLock.RUnlock()
	}
}
