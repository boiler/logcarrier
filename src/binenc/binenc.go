package binenc

import (
	"encoding/binary"
	"math"
)

// BinaryEncoder is a binary encoding facility
type BinaryEncoder struct {
	buf [8]byte
}

// New constructs new binary encoder
func New() *BinaryEncoder {
	return &BinaryEncoder{}
}

// Uleb128 performs unsigned leb128 encryption over the uint32 value
func (e *BinaryEncoder) Uleb128(value uint32) []byte {
	remaining := value >> 7
	i := 0
	for remaining != 0 {
		e.buf[i] = byte(value&0x7f | 0x80)
		i++
		value = remaining
		remaining >>= 7
	}
	e.buf[i] = byte(value & 0x7f)
	return e.buf[:i+1]
}

// Float32 maps float32 value into uint32 and then decode it in a LittleBig sequence
func (e *BinaryEncoder) Float32(value float32) []byte {
	mid := math.Float32bits(value)
	binary.LittleEndian.PutUint32(e.buf[:8], mid)
	return e.buf[:4]
}

// Float64 maps float64 value into uint64 and then decode it in a LittleBig sequence
func (e *BinaryEncoder) Float64(value float64) []byte {
	mid := math.Float64bits(value)
	binary.LittleEndian.PutUint64(e.buf[:8], mid)
	return e.buf[:8]
}

// Byte maps byte into memory
func (e *BinaryEncoder) Byte(value byte) []byte {
	e.buf[0] = value
	return e.buf[:1]
}

// Int16 into LittleBig sequence of bytes
func (e *BinaryEncoder) Int16(value int16) []byte {
	binary.LittleEndian.PutUint16(e.buf[:8], uint16(value))
	return e.buf[:2]
}

// Int32 into LittleBig sequence of bytes
func (e *BinaryEncoder) Int32(value int32) []byte {
	binary.LittleEndian.PutUint32(e.buf[:8], uint32(value))
	return e.buf[:4]
}

// Int64 into LittleBig sequence of bytes
func (e *BinaryEncoder) Int64(value int64) []byte {
	binary.LittleEndian.PutUint64(e.buf[:8], uint64(value))
	return e.buf[:8]
}

// Uint16 into LittleBig sequence of bytes
func (e *BinaryEncoder) Uint16(value uint16) []byte {
	binary.LittleEndian.PutUint16(e.buf[:8], value)
	return e.buf[:2]
}

// Uint32 into LittleBig sequence of bytes
func (e *BinaryEncoder) Uint32(value uint32) []byte {
	binary.LittleEndian.PutUint32(e.buf[:8], value)
	return e.buf[:4]
}

// Uint64 into LittleBig sequence of bytes
func (e *BinaryEncoder) Uint64(value uint64) []byte {
	binary.LittleEndian.PutUint64(e.buf[:8], value)
	return e.buf[:8]
}

// Bool into sequence of bytes
func (e *BinaryEncoder) Bool(value bool) []byte {
	if value {
		e.buf[0] = 1
	} else {
		e.buf[0] = 0
	}
	return e.buf[:1]
}
