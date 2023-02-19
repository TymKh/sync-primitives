package examples

import (
	"context"
	"sync"
	"testing"

	"async-task-manager/spike"
	"github.com/stretchr/testify/assert"
)

// Example of using simple map as cache (not recommended, but possible). Obviously map itself doesn't support TTL, but could work for semi static data
// Notice, that even though map is not thread safe, it's still safe to use it in concurrent environment, because spike.Manager is thread safe
func TestCustomCache(t *testing.T) {
	cm := make(map[string]string)
	sm := spike.NewCustomManager(spike.Handler[string]{
		Fetch: func(ctx context.Context, k string) (string, error) {
			return "value_for_" + k, nil
		},
		Set: func(k string, v string) {
			cm[k] = v
		},
		Get: func(k string) (string, bool) {
			v, ok := cm[k]
			return v, ok
		},
	})
	v, err := sm.GetResult(context.Background(), "1")
	assert.NoError(t, err)
	assert.Equal(t, v, "value_for_1")

	keys := []string{"2", "3", "4", "5", "6"}
	wg := sync.WaitGroup{}
	wg.Add(len(keys) * 101)
	for i := 0; i <= 100; i++ {
		for _, k := range keys {
			go func(k string) {
				defer wg.Done()
				res, err := sm.GetResult(context.Background(), k)

				assert.NoError(t, err)
				assert.Equal(t, res, "value_for_"+k)
			}(k)
		}
	}
	wg.Wait()

	keys = append(keys, "1")

	for _, k := range keys {
		assert.Equal(t, cm[k], "value_for_"+k)
	}
}
