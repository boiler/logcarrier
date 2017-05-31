package bindec

import (
	"encoding/binary"
	"math"
)

// ResponseReader provides data decoding from RowBinary format
type ResponseReader struct {
	src []byte
}

// New ResponseReader constructor
func New(src []byte) *ResponseReader {
	return &ResponseReader{
		src: src,
	}
}

func (rr *ResponseReader) take(b int) (res []byte, ok bool) {
	ok = len(rr.src) >= b
	if !ok {
		return nil, false
	}
	res = rr.src[:b]
	rr.src = rr.src[b:]
	return
}

// Float32 reads float32 value
func (rr *ResponseReader) Float32() (value float32, ok bool) {
	data, ok := rr.take(4)
	if !ok {
		return
	}
	mask := binary.LittleEndian.Uint32(data)
	value = math.Float32frombits(mask)
	return
}

// Float64 reads float64 value
func (rr *ResponseReader) Float64() (value float64, ok bool) {
	data, ok := rr.take(8)
	if !ok {
		return
	}
	mask := binary.LittleEndian.Uint64(data)
	value = math.Float64frombits(mask)
	return
}

// Byte reads byte value
func (rr *ResponseReader) Byte() (value byte, ok bool) {
	data, ok := rr.take(1)
	if !ok {
		return
	}
	return data[0], true
}

// Int16 reads int16 value
func (rr *ResponseReader) Int16() (value int16, ok bool) {
	data, ok := rr.take(2)
	if !ok {
		return
	}
	value = int16(binary.LittleEndian.Uint16(data))
	return
}

// Int32 reads int32 value
func (rr *ResponseReader) Int32() (value int32, ok bool) {
	data, ok := rr.take(4)
	if !ok {
		return
	}
	value = int32(binary.LittleEndian.Uint32(data))
	return
}

// Int64 reads int64 value
func (rr *ResponseReader) Int64() (value int64, ok bool) {
	data, ok := rr.take(8)
	if !ok {
		return
	}
	value = int64(binary.LittleEndian.Uint64(data))
	return
}

// Uint16 reads uint16 value
func (rr *ResponseReader) Uint16() (value uint16, ok bool) {
	data, ok := rr.take(2)
	if !ok {
		return
	}
	value = uint16(binary.LittleEndian.Uint16(data))
	return
}

// Uint32 reads uint32 value
func (rr *ResponseReader) Uint32() (value uint32, ok bool) {
	data, ok := rr.take(4)
	if !ok {
		return
	}
	value = uint32(binary.LittleEndian.Uint32(data))
	return
}

// Uint64 reads uint64 value
func (rr *ResponseReader) Uint64() (value uint64, ok bool) {
	data, ok := rr.take(8)
	if !ok {
		return
	}
	value = uint64(binary.LittleEndian.Uint64(data))
	return
}

// Bool reads boolan value
func (rr *ResponseReader) Bool() (value bool, ok bool) {
	data, ok := rr.take(1)
	if !ok {
		return
	}
	if data[0] != 0 {
		value = true
	}
	return
}

// Bytes reads n bytes from the source
func (rr *ResponseReader) Bytes(n int) (value []byte, ok bool) {
	value, ok = rr.take(n)
	if !ok {
		return
	}
	return
}

// Empty checks if the underlying buffer is empty
func (rr *ResponseReader) Empty() bool {
	return len(rr.src) == 0
}
