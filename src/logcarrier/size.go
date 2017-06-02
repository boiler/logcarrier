package main

import (
	"fmt"
	"strconv"
)

// Size is used to represent buffer size values
type Size uint64

// UnmarshalText toml unmarshalling implementation
func (s *Size) UnmarshalText(rawText []byte) error {
	text := string(rawText)
	var pos int
	for pos = 0; pos < len(text); pos++ {
		if text[pos] < '0' || text[pos] > '9' {
			break
		}
	}
	if pos == 0 {
		return fmt.Errorf("Wrong buffer size value, digits must be at the start `\033[1m%s\033[0m`", text)
	}

	digits := text[:pos]
	suffix := text[pos:]
	var factor uint64
	switch suffix {
	case "":
		factor = 1
	case "Kb":
		factor = 1024
	case "Mb":
		factor = 1024 * 1024
	case "Gb":
		factor = 1024 * 1024 * 1024
	default:
		return fmt.Errorf(
			"Unknown volume unit `\033[1m%s\033[0m` in size `\033[1m%s\033[0m`, only Kb, Mb and Gb are supported", suffix, text)
	}

	value, _ := strconv.ParseUint(digits, 10, 64)
	*s = Size(value * factor)

	return nil
}
