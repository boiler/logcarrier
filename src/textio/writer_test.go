package textio

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriter(t *testing.T) {
	buf := &bytes.Buffer{}
	data := "1\n2\n3\n456"
	w := NewWriter(buf)
	_, err := w.Write([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	if err = w.Flush(); err != nil {
		t.Fatal(err)
	}
	require.Equal(t, "1\n2\n3\n", buf.String())
	require.Equal(t, 3, w.LinesWritten())
	if err = w.FlushAll(); err != nil {
		t.Fatal(err)
	}
	require.Equal(t, data, buf.String())
}