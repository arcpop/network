package config

import (
)

var Arp struct {
    NumberOfQueueWorkers int
    RxQueueSize int
}

var Ethernet struct {
    NumberOfQueueWorkers int
    TxQueueSize, RxQueueSize int
}

var Device struct {
    RxQueueSize int
    TxQueueSize int
    RxQueueWorkers int
    TxQueueWorkers int
}

var UDP struct {
    RecvQueueSize int
    ConnectionRecvQueueSize int
}

func init()  {
    Device.RxQueueSize = 1024
    Device.TxQueueSize = 1024
    Device.RxQueueWorkers = 1
    Device.TxQueueWorkers = 1
    
    Ethernet.NumberOfQueueWorkers = 1
    Ethernet.TxQueueSize = 1024
    Ethernet.RxQueueSize = 1024
    
    Arp.NumberOfQueueWorkers = 1
    Arp.RxQueueSize = 1024
    
    UDP.RecvQueueSize = 512
    UDP.ConnectionRecvQueueSize = 512
}