package main

import (
	"os"
	"path"
)

// PathExists checks that path exists on filesystem
func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

// PathGen is a path generator with PathGen string value as a root
type PathGen string

// Join joins parts into the file path with pg as a root
func (pg PathGen) Join(parts ...string) string {
	rp := make([]string, len(parts)+1)
	rp[0] = string(pg)
	copy(rp[1:], parts)
	return path.Join(rp...)
}
