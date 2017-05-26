package main

import (
	"fmt"
	"sort"
	"strings"
)

// CompressionMethod encodes a method used for compression
type CompressionMethod int

const (
	// ZStd is used for ZStd compression
	ZStd CompressionMethod = iota

	// GZip is used for GZip compression
	GZip

	// Raw means no compression
	Raw
)

var compressionMapping = map[string]CompressionMethod{
	ZStd.String(): ZStd,
	GZip.String(): GZip,
	Raw.String():  Raw,
}

// UnsupportedError generates error message for toml encoding error
func UnsupportedError(data []byte) error {
	items := make([]string, 0, len(compressionMapping))
	for k := range compressionMapping {
		items = append(items, k)
	}
	sort.Sort(sort.StringSlice(items))
	list := strings.Join(items, ", ")
	text := string(data)
	if len(text) == 0 {
		return fmt.Errorf("no compression method given, use one of %s", list)
	}
	return fmt.Errorf("unsupported compression method `\033[1m%s\033[0m`, use one of %s", text, list)
}

// UnmarshalText toml unmarshalling implementation
func (i *CompressionMethod) UnmarshalText(text []byte) error {
	value, ok := compressionMapping[string(text)]
	if !ok {
		return UnsupportedError(text)
	}
	*i = value
	return nil
}
