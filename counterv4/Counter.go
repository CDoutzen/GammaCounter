package counterv4

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type kv struct {
	k string
	v int
}

type Counter struct {
	cache sync.Map
	mq    chan kv
	flush chan func(*sync.Map)
}

func CbFlush(m *sync.Map) {
	timestamp := strconv.Itoa(int(time.Now().Unix()))
	f, err := os.OpenFile("./cache_"+timestamp, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		log.Fatalf(err.Error())
	}
	defer f.Close()

	// Flush the old and empty

	str := strings.Builder{}
	str.WriteString("{\n")

	stream := func(key, value any) bool {
		str.WriteString(fmt.Sprintf("\t%s: %d, \n", key.(string), value.(int)))
		return true
	}
	m.Range(stream)
	str.WriteString("}")
	f.WriteString(str.String())

	if err != nil {
		log.Fatalf(err.Error())
	}
}

func (c *Counter) Init() {
	c.cache = sync.Map{}
	c.mq = make(chan kv, 10240)
	c.flush = make(chan func(*sync.Map))
	go func() {
		for {
			select {
			case x := <-c.mq:
				{
					if value, ok := c.cache.LoadOrStore(x.k, x.v); ok {
						c.cache.Store(x.k, value.(int)+x.v)
					}

				}
			case flushfunc := <-c.flush:
				{
					bk := c.cache
					c.cache = sync.Map{}
					go flushfunc(&bk)
				}
			}
		}
	}()
}

func (c *Counter) Incr(key string, inc int) {
	c.mq <- kv{key, inc}
}

func (c *Counter) Flush2Broker(interval time.Duration, FuncCbFlush func(m *sync.Map)) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		<-ticker.C
		c.flush <- FuncCbFlush
	}
}

func (c *Counter) Get(key string) (int, bool) {
	if value, ok := c.cache.Load(key); ok {
		return value.(int), ok
	} else {
		return -1, ok
	}
}
