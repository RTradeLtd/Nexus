package delegator

import (
	"net/http/httputil"
	"testing"
	"time"
)

func Test_cache_Get(t *testing.T) {
	type fields struct {
		expire time.Duration
	}
	type args struct {
		put string
		get string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{"nonexistent key", fields{time.Minute}, args{"", "key"}, false},
		{"expired key", fields{0}, args{"key", "key"}, false},
		{"ok key", fields{time.Minute}, args{"key", "key"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newCache(tt.fields.expire, 5*time.Minute)
			c.Cache(tt.args.put, &httputil.ReverseProxy{})
			time.Sleep(time.Microsecond)
			if got := c.Get(tt.args.get); (got == nil) == tt.want {
				t.Errorf("cache.Exists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_cache_cleaner(t *testing.T) {
	c := newCache(0, time.Microsecond)

	// cache and wait for collector to pick up
	c.Cache("some_key", nil)
	time.Sleep(time.Millisecond)

	// check that collector picked up key
	c.mux.RLock()
	if _, f := c.store["some_key"]; f {
		t.Error("item was not removed by cleaner")
	}
	c.mux.RUnlock()

	// stop
	c.stop <- true
	time.Sleep(time.Millisecond)
}
