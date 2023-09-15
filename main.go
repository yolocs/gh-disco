package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/abcxyz/pkg/cli"
	"github.com/yolocs/gh-disco/commands"
)

func main() {
	ctx, done := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer done()

	if err := rootCmd.Run(ctx, os.Args[1:]); err != nil {
		done()
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

var rootCmd = &cli.RootCommand{
	Name:    "gh-disco",
	Version: "dev",
	Commands: map[string]cli.CommandFactory{
		"sso": func() cli.Command {
			return &commands.SSOCommand{}
		},
	},
}
