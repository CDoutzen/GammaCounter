package counterSeq

import (
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type kv struct {
	k string
	v int
}

var mux sync.Mutex

type Counter struct {
	cache   map[string]int
	mq      chan kv
	getKey  chan string
	retVal  chan int
	retNone chan bool
	flush   chan func(*map[string]int)
}

func CbFlush(m *map[string]int) {
	timestamp := strconv.Itoa(int(time.Now().Unix()))
	f, err := os.OpenFile("./cache_"+timestamp, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		log.Fatalf(err.Error())
	}
	defer f.Close()
	// callTimes := 0
	// for _, v := range *m {
	// 	callTimes += v
	// }
	// fmt.Printf("write and read %d times in 1s\n", callTimes)
	// Flush the old and empty
	buf, err := json.Marshal(m)
	if err != nil {
		log.Fatalf(err.Error())
	}
	_, err = f.Write(buf)
	if err != nil {
		log.Fatalf(err.Error())
	}
}

func Flush2Null(m *map[string]int) {

}

func (c *Counter) Init() {
	c.cache = map[string]int{}
	c.mq = make(chan kv, 2048)
	c.getKey = make(chan string)
	c.retVal = make(chan int)
	c.retNone = make(chan bool)
	c.flush = make(chan func(*map[string]int))
	go func() {
		for {
			select {
			case x := <-c.mq:
				{
					c.cache[x.k] += x.v
				}
			case flushfunc := <-c.flush:
				{
					mux.Lock()
					bk := c.cache
					c.cache = map[string]int{}
					mux.Unlock()
					go flushfunc(&bk)
				}
			case key := <-c.getKey:
				{
					if value, ok := c.cache[key]; ok {
						c.retVal <- value
					} else {
						c.retNone <- true
					}
				}
			}
		}
	}()
}

func (c *Counter) Incr(key string, inc int) {
	c.mq <- kv{key, inc}
}

func (c *Counter) Flush2Broker(interval time.Duration, FuncCbFlush func(*map[string]int)) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		<-ticker.C
		c.flush <- FuncCbFlush
	}
}

func (c *Counter) Get(key string) (int, bool) {
	c.getKey <- key
	select {
	case x := <-c.retVal:
		return x, true
	case <-c.retNone:
		return -1, false
	}
}

func TestCounterSeq() {
	// fmt.Println("In CounterS:")
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
