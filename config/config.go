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

var Tap struct {
    NumberOfQueues int
}


func init()  {
    Tap.NumberOfQueues = 1
}