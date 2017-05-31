package logio

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

func TestWriterChunks(t *testing.T) {
	buf := &bytes.Buffer{}
	data := "1\n2\n3\n456"
	w := NewWriterSize(buf, 1)
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

func TestWriterChunks2(t *testing.T) {
	buf := &bytes.Buffer{}
	data := "1\n2\n3\n456"
	w := NewWriterSize(buf, 2)
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

func TestInfty(t *testing.T) {
	buf := &bytes.Buffer{}
	data := []byte("12345\n")
	w := NewBlowingWriter(buf, 1024)
	for i := 0; i < 6000000; i++ {
		_, err := w.Write(data)
		if err != nil {
			t.Fatal(err)
		}
	}
	require.Equal(t, 0, buf.Len())
	if err := w.Flush(); err != nil {
		t.Fatal(err)
	}
	require.Equal(t, 36000000, buf.Len())
}
