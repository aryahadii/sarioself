package main

import (
	"github.com/aryahadii/sarioself/configuration"
	"github.com/aryahadii/sarioself/telegram"
	"github.com/spf13/cobra"
)

var (
	startCmd = &cobra.Command{
		Use:   "start",
		Short: "Start bot",
		Run:   start,
	}
)

func init() {
	rootCmd.AddCommand(startCmd)
}

func start(cmd *cobra.Command, args []string) {
	if configuration.SarioselfConfig.GetBool("bots.telegram.enabled") {
		telegram.StartBot()
	}
}
