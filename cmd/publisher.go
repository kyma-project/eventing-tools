package cmd

import (
	"github.com/spf13/cobra"

	loadtestpublisher "github.com/kyma-project/eventing-tools/internal/loadtest/publisher"
)

var publisherPort int

// publisherCmd represents the publisher command
var publisherCmd = &cobra.Command{
	Use:   "publisher",
	Short: "Publish events compatible with kyma eventing",
	Run: func(cmd *cobra.Command, args []string) {
		loadtestpublisher.Start(publisherPort)
	},
}

func init() {
	publisherCmd.Flags().IntVarP(&publisherPort, "listen-port", "p", 8888, "listen on port (health check, control commands)")
	rootCmd.AddCommand(publisherCmd)
}
