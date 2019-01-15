package main

import (
	"fmt"
	_ "github.com/johnharris85/pokcli/pkg/client"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var pokcliCmd = &cobra.Command{
	Use:   "pokcli",
	Short: "Pocket CLI",
}

var logLevel int

// PokcliVersion is the release TAG
var PokcliVersion string

// PokcliBuild is the current GIT commit
var PokcliBuild string

func init() {
	// Global flag across all subcommands
	pokcliCmd.PersistentFlags().IntVar(&logLevel, "logLevel", 4, "Set the logging level [0=panic, 3=warning, 5=debug]")
	pokcliCmd.AddCommand(pokcliVersion)
}

// Execute - starts the command parsing process
func main() {
	log.SetLevel(log.Level(logLevel))
	if err := pokcliCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var pokcliVersion = &cobra.Command{
	Use:   "version",
	Short: "Version and Release information about the Pokcli tool",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Pokcli, a command-line interface for Pocket")
		fmt.Printf("Version:  %s\n", PokcliVersion)
		fmt.Printf("Build:    %s\n", PokcliBuild)
	},
}
