package counter

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var mux sync.Mutex

type Counter struct {
	M *sync.Map
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
	c.M = &sync.Map{}
}

func (c *Counter) Incr(key interface{}, inc interface{}) {
	if old, loaded := c.M.LoadOrStore(key, inc); loaded {
		c.M.Store(key, old.(int)+inc.(int))
	}
}

func (c *Counter) Flush2Broker(interval time.Duration, FuncCbFlush func(*sync.Map)) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		<-ticker.C

		// for every {interval} time, lock the map and flush
		mux.Lock()
		m := c.M
		c.Init()
		mux.Unlock()
		// fmt.Println("New: ", c.M)
		// fmt.Println("Old: ", m)
		CbFlush(m)
	}
}

func (c *Counter) Get(key any) (int, bool) {
	if value, ok := c.M.Load(key); ok {
		return value.(int), ok
	} else {
		return -1, ok
	}
}
