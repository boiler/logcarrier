/*
package bufferer provides
*/

package bufferer

// Bufferer is an interface that groups set of methods that are
// needed for files syncing and log rotating
type Bufferer interface {
	// Write writes data from p to the underlying buffer
	Write(p []byte) (n int, err error)

	// Close flushes all buffered data and close the file
	Close() error

	// Flush flushes all buffered data
	Flush() error
}
