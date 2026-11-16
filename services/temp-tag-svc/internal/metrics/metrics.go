// Package metrics exposes Prometheus text-format counters without external deps.
package metrics

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"
)

type Registry struct {
	mu sync.Mutex
	c  map[string]float64
}

func New() *Registry { return &Registry{c: map[string]float64{}} }

// Inc increments name{labels...}; labels are key, value pairs.
func (r *Registry) Inc(name string, labels ...string) {
	var b strings.Builder
	b.WriteString(name)
	if len(labels) > 0 {
		b.WriteString("{")
		for i := 0; i+1 < len(labels); i += 2 {
			if i > 0 {
				b.WriteString(",")
			}
			fmt.Fprintf(&b, "%s=%q", labels[i], labels[i+1])
		}
		b.WriteString("}")
	}
	r.mu.Lock()
	r.c[b.String()]++
	r.mu.Unlock()
}

func (r *Registry) Write(w io.Writer) {
	r.mu.Lock()
	defer r.mu.Unlock()
	keys := make([]string, 0, len(r.c))
	for k := range r.c {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(w, "%s %g\n", k, r.c[k])
	}
}
