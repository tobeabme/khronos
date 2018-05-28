package khronos

import "sync"

type Counter struct {
	mux       sync.RWMutex
	Processor map[string]map[string]int
}

var Gcounter *Counter

func init() {
	Gcounter = &Counter{Processor: make(map[string]map[string]int)}
}
func NewCounter() *Counter {
	return Gcounter
}

func (c *Counter) Plus(nodeName string, quotaName string) {
	c.mux.Lock()
	defer c.mux.Unlock()

	if quotaName != "" {
		if len(c.Processor) == 0 {
			q := make(map[string]int)
			q[quotaName] = c.Processor[nodeName][quotaName] + 1
			c.Processor[nodeName] = q
		} else {
			c.Processor[nodeName][quotaName] += 1
		}
	}

}

func (c *Counter) Minus(nodeName string, quotaName string) {
	c.mux.Lock()
	defer c.mux.Unlock()

	if quotaName != "" {
		if v, ok := c.Processor[nodeName]; ok {
			if _, ok := v[quotaName]; ok {
				if c.Processor[nodeName][quotaName] > 0 {
					c.Processor[nodeName][quotaName] -= 1
				}

			}
		}

	}
}

func (c *Counter) Get(nodeName string, quotaName string) int {
	c.mux.Lock()
	defer c.mux.Unlock()

	if quotaName != "" {
		if v, ok := c.Processor[nodeName]; ok {
			if _, ok := v[quotaName]; ok {
				return c.Processor[nodeName][quotaName]
			}
		}
	}

	return 0
}
