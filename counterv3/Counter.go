package counterv3

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

type kv struct {
	k string
	v int
}

var rmux sync.RWMutex

type Counter struct {
	cache map[string]int
	mq    chan kv
	flush chan func(*map[string]int)
}

func CbFlush(m *map[string]int) {
	timestamp := strconv.Itoa(int(time.Now().Unix()))
	f, err := os.OpenFile("./cache_"+timestamp, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		log.Fatalf(err.Error())
	}
	defer f.Close()

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

func (c *Counter) Init() {
	c.cache = map[string]int{}
	c.mq = make(chan kv, 2048)
	c.flush = make(chan func(*map[string]int))
	go func() {
		for {
			select {
			case x := <-c.mq:
				{
					rmux.RLock()
					c.cache[x.k] += x.v
					rmux.RUnlock()
				}
			case flushfunc := <-c.flush:
				{
					rmux.Lock()
					bk := c.cache
					c.cache = map[string]int{}
					rmux.Unlock()
					go flushfunc(&bk)
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
	rmux.Lock()
	defer rmux.Unlock()
	if value, ok := c.cache[key]; ok {
		return value, ok
	} else {
		return -1, ok
	}
}
