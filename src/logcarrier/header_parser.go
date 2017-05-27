/*
* THE FILE WAS GENERATED WITH logparsergen --source=src/logcarrier/parsings.script --package=main
* DO NOT TOUCH IT!
 */
package main

import (
	"bytes"
	"fmt"
)

type header struct {
	rest    []byte
	Command []byte
	Key     []byte
	Group   []byte
	Dirname []byte
	Logname []byte
}

func (p *header) Parse(line []byte) (bool, error) {
	p.rest = line
	var pos int
	if pos = bytes.IndexByte(p.rest, ' '); pos < 0 {
		return false, fmt.Errorf("`[1m%s[0m` is not a prefix of \033[1m%s\033[0m", string(' '), string(p.rest))
	}
	p.Command = p.rest[:pos]
	p.rest = p.rest[pos+1:]
	if pos = bytes.IndexByte(p.rest, ' '); pos < 0 {
		return false, fmt.Errorf("`[1m%s[0m` is not a prefix of \033[1m%s\033[0m", string(' '), string(p.rest))
	}
	p.Key = p.rest[:pos]
	p.rest = p.rest[pos+1:]
	if pos = bytes.IndexByte(p.rest, ' '); pos < 0 {
		return false, fmt.Errorf("`[1m%s[0m` is not a prefix of \033[1m%s\033[0m", string(' '), string(p.rest))
	}
	p.Group = p.rest[:pos]
	p.rest = p.rest[pos+1:]
	if pos = bytes.IndexByte(p.rest, ' '); pos < 0 {
		return false, fmt.Errorf("`[1m%s[0m` is not a prefix of \033[1m%s\033[0m", string(' '), string(p.rest))
	}
	p.Dirname = p.rest[:pos]
	p.rest = p.rest[pos+1:]
	p.Logname = p.rest
	p.rest = p.rest[len(p.rest):]
	if len(p.rest) > 0 {
		return false, fmt.Errorf("Expected for line to be read out, still has `[1m%s[0m` to read", string(p.rest))
	}
	return true, nil
}
