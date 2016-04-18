//Package tap is used to create a TAP adapter with one or several queues.
package tap

//Adapter represents a tap adapter with one or several processing queues
type Adapter struct {
	queues []Tap
	stop   chan bool
    sendQueue, receiveQueue chan []byte
}

//Tap specifies the necessary methods which are needed to use a tap interface
type Tap interface {
	DoRead(receiveQueue chan []byte, stop chan bool)
	DoWrite(sendQueue chan []byte, stop chan bool)
	Read(b []byte) (err error)
	Write(b []byte) (err error)
	Close() (err error)
}

//StartProcessing sets the queues into processing mode. They will start processing packets.
func (ta *Adapter) StartProcessing(sendQueue, receiveQueue chan []byte) {
    ta.sendQueue = sendQueue
    ta.receiveQueue = receiveQueue
    ta.stop = make(chan bool)
	for i := 0; i < len(ta.queues); i++ {
		go ta.queues[i].DoWrite(ta.sendQueue, ta.stop)
		go ta.queues[i].DoRead(ta.receiveQueue, ta.stop)
	}
}

//StopProcessing will make the queues stop processing.
func (ta *Adapter) StopProcessing() {
	close(ta.stop)
}


//Close will cleanup all ressources.
func (ta *Adapter) Close() {
	if ta.stop != nil {
		ta.StopProcessing()
	}
	for i := 0; i < len(ta.queues); i++ {
		ta.queues[i].Close()
	}
}
