package khronos

import (
	"testing"
	"time"
)

func TestAgentCommandRun(t *testing.T) {
	shutdownCh := make(chan struct{})
	defer close(shutdownCh)

	a := &Agent{
		ShutdownCh: shutdownCh,
	}

	a.Run()

	time.Sleep(2 * time.Second)

	// Verify it runs "forever"
	select {}

}
