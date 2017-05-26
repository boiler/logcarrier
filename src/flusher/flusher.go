package flusher

// Flusher interface has two methods:
//  regular buffer flushing
//  checks if forced flushing makes a sense
type Flusher interface {
	Flush() error
	WorthFlushing() bool
}
