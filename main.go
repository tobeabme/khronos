package main

import (
	"fmt"
	"os"

	"github.com/mitchellh/cli"
	"github.com/tobeabme/khronos/khronos"
)

func main() {
	c := cli.NewCLI("khronos", "1.0.0")
	c.Args = os.Args[1:]
	c.HelpFunc = cli.BasicHelpFunc("khronos")

	ui := &cli.BasicUi{Writer: os.Stdout}

	c.Commands = map[string]cli.CommandFactory{
		"agent": func() (cli.Command, error) {
			return &khronos.Command{
				Ui: ui,
			}, nil
		},
	}

	exitStatus, err := c.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing CLI: %s\n", err.Error())
	}
	os.Exit(exitStatus)
}
