package cheapbuf

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScanner(t *testing.T) {
	s := NewScanner(NewReaderSize(1))
	sample := []string{"1", "2", "34", "56", "789", "abc"}
	reader := strings.NewReader(strings.Join(sample, "\n") + "\n")
	s.SetReader(reader)

	i := 0
	for s.Scan() {
		require.Equal(t, sample[i]+"\n", string(s.Bytes()))
		i++
	}
	if s.Err() != nil {
		t.Fatal(s.Err())
	}
}

func TestScannerLargeBuffer(t *testing.T) {
	s := NewScanner(NewReader())
	sample := []string{"1", "2", "34", "56", "789", "abc"}
	reader := strings.NewReader(strings.Join(sample, "\n") + "\n")
	s.SetReader(reader)

	i := 0
	for s.Scan() {
		require.Equal(t, sample[i]+"\n", string(s.Bytes()))
		i++
	}
	if s.Err() != nil {
		t.Fatal(s.Err())
	}
}
