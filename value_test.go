// Copyright (c) 2015 Datacratic. All rights reserved.

package jq

import (
	"encoding/json"
	"sync"
	"testing"
)

func TestParseJSON(t *testing.T) {
	n := len(samples)

	for i := 0; i < n; i++ {
		sample := samples[i]

		v := Value{}
		if err := v.Unmarshal(sample); err != nil {
			t.Fatal(err)
		}
	}
}

func BenchmarkParseGoJSON(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m := make(map[string]interface{})
		if err := json.Unmarshal(samples[i%len(samples)], &m); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseJSON(b *testing.B) {
	valuePool := sync.Pool{
		New: func() interface{} { return &Value{} },
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := valuePool.Get().(*Value)
		if err := v.Unmarshal(samples[i%len(samples)]); err != nil {
			b.Fatal(err)
		}

		valuePool.Put(v)
	}
}
