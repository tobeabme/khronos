package khronos

import (
	"fmt"
	"testing"
)

//go test -v -run=TestCounter
func TestCounter(t *testing.T) {
	nodeName := "192.168.1.11"
	c := NewCounter()
	c.Plus(nodeName, "undone")
	c.Plus(nodeName, "undone")
	c.Plus(nodeName, "websocket")
	c.Minus(nodeName, "undone")
	c.Minus(nodeName, "undone")
	c.Minus(nodeName, "undone")

	fmt.Println("Counter.Plus++++++++", c, Count)

}
