/*
package bufferer provides
*/

package bufferer

import (
	"bindec"
	"binenc"
	"bytes"
)

// Bufferer is an interface that groups set of methods that are
// needed for files syncing and log rotating
type Bufferer interface {
	// Write writes data from p to the underlying buffer
	Write(p []byte) (n int, err error)

	// Close flushes all buffered data and close the file
	Close() error

	// Flush flushes all buffered data
	Flush() error

	// PostWrite this differs for different purposes. Say, this will flush
	// the front buffer for the zstd and will do nothing for raw
	PostWrite() error

	// Logrotate rotates underlying log
	Logrotate(newname string) error

	// DumpState dumps the state of the bufferer object
	DumpState(enc *binenc.Encoder, dest *bytes.Buffer)

	// RestoreState restores the state of the bufferer object
	RestoreState(src *bindec.Decoder)
}
