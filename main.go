package main

import (
	"interview/gamma/counterv5"
	"math/rand"
	"runtime"
	"time"
)

func main() {
	cnter := counterv5.Counter{}
	cnter.Init()
	// wp := sync.WaitGroup{}
	thread := 5
	// wp.Add(thread)
	runtime.GOMAXPROCS(thread + 1)
	for j := 0; j < thread; j++ {
		key1 := string(byte(rand.Uint32() % 256))
		key2 := string(byte(rand.Uint32() % 256))
		go func(incKey, getKey string) {
			for i := 1; i <= 100000000; i++ {
				// time.Sleep(time.Millisecond * 20)
				cnter.Incr(incKey, 1)
				cnter.Get(getKey)
			}
			// wp.Done()
		}(key1, key2)

	}

	go func() {
		cnter.Flush2Broker(time.Second*1, counterv5.CbFlush)
		// wp.Done()
	}()
	ticker := time.After(time.Second * 5)
	// for {
	// 	select {
	// 	case <-ticker:
	// 		return
	// 	default:
	// 		continue
	// 	}
	// }
	<-ticker
	// wp.Wait()
}
