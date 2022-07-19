package main

import "interview/gamma/counterSeq"

// counterRW uses sync.Map to assure concurrent-safe for reading and writing \
// but dirty read, phatom read and non-repetitive read  may happen as incr operation for sync.Map is not atomic.
// counterSeq uses common map[string]int to save k-v, the Incr, Get request is sent to a queue to prevent the situations above \
// all the operations are handled in sequence.

func main() {
	// counterRW.TestCounterRW()
	counterSeq.TestCounterSeq()
}
