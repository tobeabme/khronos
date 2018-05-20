package khronos

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/mitchellh/cli"
)

const (
	// gracefulTimeout controls how long we wait before forcefully terminating
	gracefulTimeout = 3 * time.Second
)

// Command run khronos agent
type Command struct {
	Ui         cli.Ui
	ShutdownCh <-chan struct{}

	config *Configuration
	agent  *Agent
}

// Help returns agent command usage to the CLI.
func (c *Command) Help() string {
	helpText := `
Usage: khronos agent [options]
	Run khronos agent
Options:
  -e=prod						  The environment you want to run in (local, dev, sit, prod)
  -bind-ip=0.0.0.0		          Address to bind network listeners to.
  -bind-port=10001        		  Address to bind network listeners to.
  -node=hostname                  Name of this node. Must be unique in the cluster
  -backend=[etcd|consul|zk|redis] Backend storage to use, etcd, consul, zk (zookeeper) or redis.
                                  The default is etcd.
  -backend-machine=127.0.0.1:2379 Backend storage servers addresses to connect to. This flag can be
                                  specified multiple times.
  -rpc-port=10005                 RPC Port used to communicate with clients. Only used when server.
                                  The RPC IP Address will be the same as the bind address.
  -mail-host                      Mail server host address to use for notifications.
  -mail-port                      Mail server port.
  -mail-username                  Mail server username used for authentication.
  -mail-password                  Mail server password to use.
  -mail-from                      From email address to use.
  -log-level=info                 Log level (debug, info, warn, error, fatal, panic). Default to info.
`
	return strings.TrimSpace(helpText)
}

// Synopsis returns the purpose of the command for the CLI
func (c *Command) Synopsis() string {
	return "Run khronos"
}

func (c *Command) Run(args []string) int {
	//be compatible with goptions in config
	os.Args = os.Args[1:]

	c.config = NewConfig()

	agent := NewAgent(c.config)
	if err := agent.Start(); err != nil {
		c.Ui.Error(err.Error())
		return 1
	}
	c.agent = agent

	return c.handleSignals()
}

// handleSignals blocks until we get an exit-causing signal
func (c *Command) handleSignals() int {
	signalCh := make(chan os.Signal, 4)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

WAIT:
	// Wait for a signal
	var sig os.Signal
	select {
	case s := <-signalCh:
		sig = s
	case <-c.ShutdownCh:
		sig = os.Interrupt
	}
	c.Ui.Output(fmt.Sprintf("Caught signal: %v", sig))

	// Check if this is a SIGHUP
	if sig == syscall.SIGHUP {
		c.handleReload()
		goto WAIT
	}

	// Check if we should do a graceful leave
	graceful := false
	if sig == syscall.SIGTERM || sig == os.Interrupt {
		graceful = true
	}

	// Fail fast if not doing a graceful leave
	if !graceful {
		return 1
	}

	// Attempt a graceful leave
	gracefulCh := make(chan struct{})
	c.Ui.Output("Gracefully shutting down agent...")
	log.Info("agent: Gracefully shutting down agent...")
	go func() {
		if err := c.agent.Leave(); err != nil {
			c.Ui.Error(fmt.Sprintf("Error: %s", err))
			log.Error(fmt.Sprintf("Error: %s", err))
			return
		}
		close(gracefulCh)
	}()

	// Wait for leave or another signal
	select {
	case <-signalCh:
		return 1
	case <-time.After(gracefulTimeout):
		return 1
	case <-gracefulCh:
		return 0
	}
}

// handleReload is invoked when we should reload our configs, e.g. SIGHUP
func (c *Command) handleReload() {
	c.Ui.Output("Reloading configuration...")
	newConf := ReadConfig()
	if newConf == nil {
		c.Ui.Error(fmt.Sprintf("Failed to reload configs"))
		return
	}
	c.config = newConf

}
