package bindec

import (
	"encoding/binary"
	"math"
)

// Decoder provides data decoding from RowBinary format
type Decoder struct {
	src []byte
}

// New ResponseReader constructor
func New(src []byte) *Decoder {
	return &Decoder{
		src: src,
	}
}

// SetSource sets the source to the given data.
func (rr *Decoder) SetSource(src []byte) {
	rr.src = src
}

func (rr *Decoder) take(b int) (res []byte, ok bool) {
	ok = len(rr.src) >= b
	if !ok {
		return nil, false
	}
	res = rr.src[:b]
	rr.src = rr.src[b:]
	return
}

// Float32 reads float32 value
func (rr *Decoder) Float32() (value float32, ok bool) {
	data, ok := rr.take(4)
	if !ok {
		return
	}
	mask := binary.LittleEndian.Uint32(data)
	value = math.Float32frombits(mask)
	return
}

// Float64 reads float64 value
func (rr *Decoder) Float64() (value float64, ok bool) {
	data, ok := rr.take(8)
	if !ok {
		return
	}
	mask := binary.LittleEndian.Uint64(data)
	value = math.Float64frombits(mask)
	return
}

// Byte reads byte value
func (rr *Decoder) Byte() (value byte, ok bool) {
	data, ok := rr.take(1)
	if !ok {
		return
	}
	return data[0], true
}

// Int16 reads int16 value
func (rr *Decoder) Int16() (value int16, ok bool) {
	data, ok := rr.take(2)
	if !ok {
		return
	}
	value = int16(binary.LittleEndian.Uint16(data))
	return
}

// Int32 reads int32 value
func (rr *Decoder) Int32() (value int32, ok bool) {
	data, ok := rr.take(4)
	if !ok {
		return
	}
	value = int32(binary.LittleEndian.Uint32(data))
	return
}

// Int64 reads int64 value
func (rr *Decoder) Int64() (value int64, ok bool) {
	data, ok := rr.take(8)
	if !ok {
		return
	}
	value = int64(binary.LittleEndian.Uint64(data))
	return
}

// Uint16 reads uint16 value
func (rr *Decoder) Uint16() (value uint16, ok bool) {
	data, ok := rr.take(2)
	if !ok {
		return
	}
	value = uint16(binary.LittleEndian.Uint16(data))
	return
}

// Uint32 reads uint32 value
func (rr *Decoder) Uint32() (value uint32, ok bool) {
	data, ok := rr.take(4)
	if !ok {
		return
	}
	value = uint32(binary.LittleEndian.Uint32(data))
	return
}

// Uint64 reads uint64 value
func (rr *Decoder) Uint64() (value uint64, ok bool) {
	data, ok := rr.take(8)
	if !ok {
		return
	}
	value = uint64(binary.LittleEndian.Uint64(data))
	return
}

// Bool reads boolan value
func (rr *Decoder) Bool() (value bool, ok bool) {
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
func (rr *Decoder) Bytes(n int) (value []byte, ok bool) {
	value, ok = rr.take(n)
	if !ok {
		return
	}
	return
}

// Empty checks if the underlying buffer is empty
func (rr *Decoder) Empty() bool {
	return len(rr.src) == 0
}
