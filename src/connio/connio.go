package connio

import (
	"net"
	"time"
)

// Reader that puts deadline on to the socket read. Can be reused for
// different underlying net.Conn-s
type Reader struct {
	conn  net.Conn
	await time.Duration
}

// NewReader constructor
func NewReader(await time.Duration) *Reader {
	return &Reader{
		conn:  nil,
		await: await,
	}
}

// SetConn sets underlying connection
func (r *Reader) SetConn(conn net.Conn) {
	r.conn = conn
}

// Read implementation
func (r *Reader) Read(p []byte) (n int, err error) {
	r.conn.SetReadDeadline(time.Now().Add(r.await))
	return r.Read(p)
}
