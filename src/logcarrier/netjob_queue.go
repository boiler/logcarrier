package main

import (
	"net"
)

type tailed struct {
	name string
	conn net.Conn
}

// NetJobQueue is a queue for sockets what are expecting to feed us with tailed data
type NetJobQueue struct {
	queue chan tailed
}

// NewNetJobQueue constructor
func NewNetJobQueue(size int) *NetJobQueue {
	if size < 0 {
		panic("Queue size must be 0 or greater")
	}
	return &NetJobQueue{
		queue: make(chan tailed, size),
	}
}

// Get retrieves element from the queue
func (tdq *NetJobQueue) Get() (name string, conn net.Conn) {
	item := <-tdq.queue
	return item.name, item.conn
}

// Put puts element into the queue
func (tdq *NetJobQueue) Put(name string, conn net.Conn) {
	tdq.queue <- tailed{
		name: name,
		conn: conn,
	}
}
