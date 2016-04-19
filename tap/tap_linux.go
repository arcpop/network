// +build linux

package tap

import (
	"io"
	"log"
	"net"
	"os"
	"syscall"
	"unsafe"
	"github.com/arcpop/network/util"
	"github.com/arcpop/network/config"
)

type tapQueue struct {
	file   *os.File
	parent *Adapter
}


//NewAdapter creates a new Tap adapter with the corresponding queues for processing
func NewAdapter(name string) (ta *Adapter, err error) {
	ta = &Adapter{ name: name }
	ta.queues = make([]Tap, config.Tap.NumberOfQueues)

	for i := 0; i < config.Tap.NumberOfQueues; i++ {
		var f *os.File
		f, err = createTap(name)
		if i == 0 && err != nil {
			var iface *net.Interface
			iface, err = net.InterfaceByName(ta.name)
			if err == nil {
				ta.mtu = iface.MTU
			}
		}
		if err != nil {
			for j := i - 1; j >= 0; j-- {
				ta.queues[j].Close()
			}
			return nil, err
		}
		ta.queues[i] = &tapQueue{file: f, parent: ta}
	}

	return ta, nil
}

func (tq *tapQueue) Close() error {
	f := tq.file
	tq.file = nil
	return f.Close()
}

func (tq *tapQueue) Read(b []byte) (err error) {
	_, err = io.ReadFull(tq.file, b)
	return err
}

func (tq *tapQueue) Write(b []byte) (err error) {
	n := 0
	for n < len(b) {
		var i int
		i, err = tq.file.Write(b)
		if err != nil {
			return
		}
		n += i
	}
	return
}

func (tq *tapQueue) DoRead(receiveQueue chan []byte, stop chan bool) {
	buf := make([]byte, tq.parent.mtu)
	for {
		if util.ChannelClosed(stop) {
			return
		}
		err := tq.Read(buf)
		if err != nil {
			log.Println("Tap: Failed to read a packet", err)
		}
		receiveQueue <- buf
		buf = make([]byte, tq.parent.mtu)
	}
}

func (tq *tapQueue) DoWrite(sendQueue chan []byte, stop chan bool) {
	for {
		buf, ok := <-sendQueue
		if !ok {
			return
		}
		if len(buf) > tq.parent.mtu {
			log.Println("Tap: Writing a too long packet to the wire:", len(buf), tq.parent.mtu)
		}
		err := tq.Write(buf)
		if err != nil {
			log.Println("Tap: Failed to write a packet", err)
		}
	}
}

func createTap(name string) (*os.File, error) {
	const (
		IFF_MULTI_QUEUE int = 0x0100
	)
	type ifreq struct {
		ifr_name  [syscall.IFNAMSIZ]byte
		ifr_flags uint16
	}

	f, err := os.OpenFile("/dev/net/tun", syscall.O_RDWR, 0)
	if err != nil {
		return nil, err
	}

	ifr := &ifreq{ifr_flags: uint16(syscall.IFF_TAP | IFF_MULTI_QUEUE)}
	copy(ifr.ifr_name[:], []byte(name))

	err = ioctl(f.Fd(), syscall.TUNSETIFF, uintptr(unsafe.Pointer(ifr)))
	if err != nil {
		f.Close()
		return nil, err
	}
	return f, nil
}

func ioctl(fd, cmd, ptr uintptr) error {
	_, _, e := syscall.Syscall(syscall.SYS_IOCTL, fd, cmd, ptr)
	if e != 0 {
		return e
	}
	return nil
}

func (ta *Adapter) GetMTU() int {
	return ta.mtu
	
}

func (ta *Adapter) GetHWAddr() net.HardwareAddr {
	var buf [6]byte
	copy(buf[:], ta.hwAddr)
	return net.HardwareAddr(buf[:])
}