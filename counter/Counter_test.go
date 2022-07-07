package counter

import "testing"

func TestCounter(t *testing.T) {
	counter := Counter{}
	counter.Init()
	for i := 0; i < 1000; i++ {
		counter.Incr("a", 1)
		if value, ok := counter.Get("a"); ok {
			t.Logf("Call %d times: %d\n", i, value)
		}
	}
}
