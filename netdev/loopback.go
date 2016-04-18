package netdev

import (
	"sync/atomic"
)

var loopbackQueue = make(chan []byte, 1024)

func loopbackRxPacket(dev *NetDev) []byte {
    pkt := <- loopbackQueue
    atomic.AddUint64(&dev.Stats.RxBytes, uint64(len(pkt)))
    atomic.AddUint64(&dev.Stats.RxPackets, 1)
    return pkt
}

func loopbackTxPacket(dev *NetDev, pkt []byte) {
    l := uint64(len(pkt))
    select {
        case loopbackQueue <- pkt:
            atomic.AddUint64(&dev.Stats.TxBytes, l)
            atomic.AddUint64(&dev.Stats.TxPackets, 1)
        default:
            atomic.AddUint64(&dev.Stats.TxErrors, 1)
    }
}