package main

import (
	"fmt"
	"unsafe"
)

// LogrotateMethod describes methods of log rotation
type LogrotateMethod int

const (
	// LogrotatePeriodic means to rely on periodic automatic log
	// rotation
	LogrotatePeriodic LogrotateMethod = iota

	// LogrotateGuided means to rely on guided log rotation via
	// the carrier protoocol
	LogrotateGuided

	// LogrotateBoth means both periodic and guided log rotations
	// methods are allowed
	LogrotateBoth
)

func (lm LogrotateMethod) String() string {
	switch lm {
	case LogrotatePeriodic:
		return "periodic"
	case LogrotateGuided:
		return "guided"
	case LogrotateBoth:
		return "both"
	default:
		panic(fmt.Errorf("unsupported log rotation method %d", lm))
	}
}

// UnmarshalText toml unmarshalling implementation
func (lm *LogrotateMethod) UnmarshalText(text []byte) error {
	choices := map[string]LogrotateMethod{
		LogrotatePeriodic.String(): LogrotatePeriodic,
		LogrotateGuided.String():   LogrotateGuided,
		LogrotateBoth.String():     LogrotateBoth,
	}
	method, ok := choices[*(*string)(unsafe.Pointer(&text))]
	if !ok {
		return fmt.Errorf("Unsupported log rotation type `\033[1m%s\033[0m`", string(text))
	}
	*lm = method
	return nil
}
