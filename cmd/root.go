/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	loadtest bool
	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:   "eventing-tools",
		Short: "A collection of tools to send and receive events in kyma",
		Long: `Eventing-tools is a collection of tools to send and receive events in kyma

The tools are meant to be run in a kubernetes cluster with kyma eventing enabled.
`,
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&loadtest, "loadtest", "l", false, "switch publisher/subscriber to loadtest mode")
}
