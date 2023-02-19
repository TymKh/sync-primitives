package examples

import (
	"context"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"async-task-manager/spike"
	"github.com/stretchr/testify/assert"
)

// Imagine scenario when you have a highload key-value storage, that stores data valid for some time
func TestSimpleKV(t *testing.T) {
	cacheControl := new(int32)
	mockReadDB := func(ctx context.Context, key string) (string, error) {
		atomic.AddInt32(cacheControl, 1)
		return "value_for_" + key + "-" + time.Now().String(), nil
	}

	sm := spike.NewManager(mockReadDB, time.Second*2)

	keys := []string{"1", "2", "3", "4", "5", "6"}
	var values []string
	for _, k := range keys {
		v, err := sm.GetResult(context.Background(), k)
		if err != nil {
			t.Errorf("error: %v", err)
			return
		}
		values = append(values, v)
	}

	//Imitate 1000 concurrent requests
	wg := sync.WaitGroup{}
	wg.Add(1000)
	for i := 0; i < 1000; i++ {
		go func(i int) {
			defer wg.Done()
			el := rand.Intn(len(keys))
			v, err := sm.GetResult(context.Background(), keys[el])
			if err != nil {
				t.Errorf("error: %v", err)
				return
			}
			if v != values[el] {
				t.Errorf("value mismatch: %v != %v", v, values[el])
			}
		}(i)
	}
	//Wait for all requests to finish
	wg.Wait()
	//Check that cache was used
	assert.Equal(t, *cacheControl, int32(6))
}
