package khronos

type Counter struct {
	Processor map[string]map[string]uint64
}

var Count *Counter

func NewCounter() *Counter {
	c := &Counter{Processor: make(map[string]map[string]uint64)}
	Count = c
	return c
}

func (c *Counter) Plus(nodeName string, quotaName string) {
	if len(c.Processor) == 0 {
		q := make(map[string]uint64)
		q[quotaName] = c.Processor[nodeName][quotaName] + 1
		c.Processor[nodeName] = q
	} else {
		c.Processor[nodeName][quotaName] += 1
	}

}

func (c *Counter) Minus(nodeName string, quotaName string) {
	if v, ok := c.Processor[nodeName]; ok {
		if _, ok := v[quotaName]; ok {
			if c.Processor[nodeName][quotaName] > 0 {
				c.Processor[nodeName][quotaName] -= 1
			}

		}
	}

}
