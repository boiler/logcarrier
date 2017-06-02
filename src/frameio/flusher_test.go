package frameio

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWorthFlushing(t *testing.T) {
	under := ioutil.Discard
	w := NewWriterSize(under, 2)

	data := []byte("12\n")
	single := []byte("1\n")

	// Write some data, no flush. Flushing is OK
	_, _ = w.Write(single)
	require.True(t, w.WorthFlushing())

	// Write some data, flush happened, flushing is pointless.
	_, _ = w.Write(data)
	require.False(t, w.WorthFlushing())

	// Write some data, flush happened, pointless again.
	_, _ = w.Write(data)
	require.False(t, w.WorthFlushing())

	// and again.
	_, _ = w.Write(data)
	require.False(t, w.WorthFlushing())

	// Write some data, no flush.
	_, _ = w.Write(single)
	// Thus flushing is OK.
	require.True(t, w.WorthFlushing())
	_, _ = w.Write(single)
	// Flush happened, flushing is pointless.
	require.False(t, w.WorthFlushing())
	// No flush happened, thus flushing makes a sense
	require.True(t, w.WorthFlushing())
}
