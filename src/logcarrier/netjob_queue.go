package main

import (
	"net"
)

// Conn represents name of the connection and the connection itself
type Conn struct {
	Name string   // Name of the file
	Conn net.Conn // Socket
	Size int      // Bytes to read (-1 for unknown)
}

// NetJobQueue is a queue for sockets what are expecting to feed us with tailed data
type NetJobQueue struct {
	queue chan Conn
}

// NewNetJobQueue constructor
func NewNetJobQueue(size int) *NetJobQueue {
	if size < 0 {
		panic("Queue size must be 0 or greater")
	}
	return &NetJobQueue{
		queue: make(chan Conn, size),
	}
}

// Get retrieves element from the queue
func (tdq *NetJobQueue) Get() chan Conn {
	return tdq.queue
}

// Put puts element into the queue
func (tdq *NetJobQueue) Put(name string, conn net.Conn, size int) {
	tdq.queue <- Conn{
		Name: name,
		Conn: conn,
		Size: size,
	}
}
