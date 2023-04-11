package cmd

import (
	"github.com/spf13/cobra"

	loadtestpublisher "github.com/kyma-project/eventing-tools/internal/loadtest/publisher"
	validationtestpublisher "github.com/kyma-project/eventing-tools/internal/validationtest/publisher"
)

var publisherPort int

// publisherCmd represents the publisher command
var publisherCmd = &cobra.Command{
	Use:   "publisher",
	Short: "Publish events compatible with kyma eventing",
	Run: func(cmd *cobra.Command, args []string) {
		switch loadtest {
		case true:
			loadtestpublisher.Start(publisherPort)
		case false:
			validationtestpublisher.Start()
		}
	},
}

func init() {
	publisherCmd.Flags().IntVarP(&publisherPort, "listen-port", "p", 8888, "listen on port (health check, control commands)")
	rootCmd.AddCommand(publisherCmd)
}
