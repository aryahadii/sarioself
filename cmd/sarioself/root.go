package main

import (
	"github.com/aryahadii/sarioself/configuration"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kanal <subcommand>",
	Short: "Telegram bot",
	Run:   nil,
}

func init() {
	cobra.OnInitialize(func() {
		configuration.LoadConfig()
	})
}
