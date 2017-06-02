package frameio

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriter(t *testing.T) {
	buf := &bytes.Buffer{}
	w := NewWriterSize(buf, 3)

	frames := []string{
		"1",
		"23",
		"456",
		"7890",
		"a",
		"abcd",
	}
	rf := make([][]byte, len(frames))
	for i, frame := range frames {
		rf[i] = []byte(frame)
	}

	if n, err := w.Write(rf[0]); err != nil {
		t.Fatal(err)
	} else {
		require.Equal(t, 1, n)
		require.Equal(t, "", buf.String())
	}

	if n, err := w.Write(rf[1]); err != nil {
		t.Fatal(err)
	} else {
		require.Equal(t, 2, n)
		require.Equal(t, "", buf.String())
	}

	if n, err := w.Write(rf[2]); err != nil {
		t.Fatal(err)
	} else {
		require.Equal(t, 3, n)
		require.Equal(t, "123", buf.String())
	}

	if n, err := w.Write(rf[3]); err != nil {
		t.Fatal(err)
	} else {
		require.Equal(t, 4, n)
		require.Equal(t, "1234567890", buf.String())
	}

	if n, err := w.Write(rf[4]); err != nil {
		t.Fatal(err)
	} else {
		require.Equal(t, 1, n)
		require.Equal(t, "1234567890", buf.String())
	}

	if n, err := w.Write(rf[5]); err != nil {
		t.Fatal(err)
	} else {
		require.Equal(t, 4, n)
		require.Equal(t, "1234567890aabcd", buf.String())
	}
}
