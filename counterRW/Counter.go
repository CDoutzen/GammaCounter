package counterRW

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type kv struct {
	k string
	v int
}

var caches_num = 5

var mux sync.Mutex

type Counter struct {
	caches []sync.Map
	mqs    []chan kv
	flush  chan func(*[]sync.Map)
}

func Flush2Null(m *[]sync.Map) {

}

func CbFlush(m *[]sync.Map) {

	// flush into file instead of broker
	timestamp := strconv.Itoa(int(time.Now().Unix()))
	f, err := os.OpenFile("./cache_"+timestamp, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		log.Fatalf(err.Error())
	}
	defer f.Close()

	// Flush the old and empty
	str := strings.Builder{}
	str.WriteString("{\n")
	callTimes := 0
	stream := func(key, value any) bool {
		str.WriteString(fmt.Sprintf("\t%s: %d, \n", key.(string), value.(int)))
		callTimes += value.(int)
		return true
	}
	for _, mi := range *m {
		mi.Range(stream)
	}
	fmt.Printf("write and read %d times in 1s\n", callTimes)
	str.WriteString("}")
	f.WriteString(str.String())

	if err != nil {
		log.Fatalf(err.Error())
	}
}

func (c *Counter) Init() {
	c.caches = make([]sync.Map, caches_num)
	c.mqs = make([]chan kv, caches_num)
	for i := range c.mqs {
		c.mqs[i] = make(chan kv, 256)
	}
	c.flush = make(chan func(*[]sync.Map))

	// Async: Goroutine for Incr queue
	for i := range c.caches {
		k := i
		go func(j int) {
			for {
				x := <-c.mqs[j]
				if value, ok := c.caches[j].LoadOrStore(x.k, x.v); ok {
					c.caches[j].Store(x.k, value.(int)+x.v)
				}
			}
		}(k)
	}
	// Async: Goroutine for Flush
	go func() {
		for {
			flushfunc := <-c.flush
			mux.Lock()
			bk := c.caches
			c.caches = make([]sync.Map, caches_num)
			mux.Unlock()
			go flushfunc(&bk)
		}
	}()
}

func (c *Counter) Incr(key string, inc int) {
	a := int(key[0]) % caches_num
	c.mqs[a] <- kv{key, inc}
}

func (c *Counter) Flush2Broker(interval time.Duration, FuncCbFlush func(m *[]sync.Map)) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		<-ticker.C
		c.flush <- FuncCbFlush
	}
}

// sync.Map assures concurrent-safe for reading
func (c *Counter) Get(key string) (int, bool) {
	if value, ok := c.caches[int(key[0]%5)].Load(key); ok {
		return value.(int), ok
	} else {
		return -1, ok
	}
}

func TestCounterRW() {
	fmt.Println("In CounterO:")
	cnter := Counter{}
	cnter.Init()
	// wp := sync.WaitGroup{}
	thread := 5
	// wp.Add(thread)
	runtime.GOMAXPROCS(thread + 1)

	for j := 0; j < thread; j++ {
		go func(incKey, getKey string) {
			for i := 1; i <= 100000000; i++ {
				// time.Sleep(time.Millisecond * 20)
				cnter.Incr(incKey, 1)
				cnter.Get(getKey)
			}
			// wp.Done()
		}(string(byte(rand.Uint32()%256)), string(byte(rand.Uint32()%256)))
	}

	go func() {
		cnter.Flush2Broker(time.Second*1, Flush2Null)
		// wp.Done()
	}()
	// wp.Wait()
	ticker := time.After(time.Millisecond * 5100)
	<-ticker
}
