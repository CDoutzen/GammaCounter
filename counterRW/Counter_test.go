package counterRW

import (
	rs "interview/gamma/randString"
	"testing"
	"time"
)

func BenchmarkCounterReadWrite(b *testing.B) {
	cnter := Counter{}
	cnter.Init()
	incKey := make([]string, 20)
	for i := 0; i < 20; i++ {
		incKey[i] = rs.RandString(7)
	}
	n := len(incKey)
	go func() {
		cnter.Flush2Broker(time.Second*1, Flush2Null)
	}()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			cnter.Incr(incKey[i%n], 1)
			cnter.Get(incKey[i%n])
			i++
		}
	})
}

func BenchmarkCounterReadOnly(b *testing.B) {
	cnter := Counter{}
	cnter.Init()
	incKey := make([]string, 20)
	for i := 0; i < 20; i++ {
		incKey[i] = rs.RandString(7)
	}
	n := len(incKey)
	go func() {
		cnter.Flush2Broker(time.Second*1, Flush2Null)
	}()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			cnter.Get(incKey[i%n])
			i++
		}
	})
}

func BenchmarkCounterWriteOnly(b *testing.B) {
	cnter := Counter{}
	cnter.Init()
	incKey := make([]string, 20)
	for i := 0; i < 20; i++ {
		incKey[i] = rs.RandString(7)
	}

	go func() {
		cnter.Flush2Broker(time.Second*1, Flush2Null)
	}()
	n := len(incKey)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			cnter.Incr(incKey[i%n], 1)
			i++
		}
	})
}
