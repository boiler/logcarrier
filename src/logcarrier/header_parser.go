/*
* THE FILE WAS GENERATED WITH logparsergen --source=parsings.script --package=main
* DO NOT TOUCH IT!
 */
package main

import (
	"bytes"
	"fmt"
	"strconv"
	"unsafe"
)

type headerParser struct {
	rest    []byte
	Command []byte
	Key     []byte
	Group   []byte
	Dirname []byte
	Logname []byte
	Size    int64
	NewName []byte
}

var commandPattern = []byte("DATA")
var rotatePattern = []byte("ROTATE")

func (p *headerParser) Parse(line []byte) (bool, error) {
	if len(line) > 0 && line[len(line)-1] == '\n' {
		line = line[:len(line)-1]
	}
	p.rest = line
	var pos int
	if pos = bytes.IndexByte(p.rest, ' '); pos < 0 {
		return false, fmt.Errorf("No Command parameter was found in `%s`", string(line))
	}
	p.Command = p.rest[:pos]
	p.rest = p.rest[pos+1:]

	if pos = bytes.IndexByte(p.rest, ' '); pos < 0 {
		return false, fmt.Errorf("No Key parameter was found in `%s`", string(line))
	}
	p.Key = p.rest[:pos]
	p.rest = p.rest[pos+1:]

	if pos = bytes.IndexByte(p.rest, ' '); pos < 0 {
		return false, fmt.Errorf("No group parameter was found in `%s`", string(line))
	}
	p.Group = p.rest[:pos]
	p.rest = p.rest[pos+1:]

	if pos = bytes.IndexByte(p.rest, ' '); pos < 0 {
		return false, fmt.Errorf("No dirname parameter was found in `%s`", string(line))
	}
	p.Dirname = p.rest[:pos]
	p.rest = p.rest[pos+1:]

	if pos = bytes.IndexByte(p.rest, ' '); pos < 0 {
		p.Logname = p.rest
		p.Size = 0
		p.rest = p.rest[len(p.rest):]
		return true, nil
	}
	p.Logname = p.rest[:pos]
	p.rest = p.rest[pos+1:]
	if len(p.rest) == 0 {
		return true, nil
	}

	var err error
	switch *(*string)(unsafe.Pointer(&p.Command)) {
	case "DATA":
		p.Size, err = strconv.ParseInt(*(*string)(unsafe.Pointer(&p.rest)), 10, 64)
		if err != nil {
			return false, fmt.Errorf("Malformed size `%s` in `%s`", string(p.rest), string(line))
		}
	case "ROTATE":
		p.NewName = p.rest
	}

	return true, nil
}
