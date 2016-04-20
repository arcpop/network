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


func init()  {
}