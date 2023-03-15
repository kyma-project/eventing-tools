package cmd

import (
	"github.com/spf13/cobra"

	loadtestpublisher "github.com/kyma-project/eventing-tools/internal/loadtest/publisher"
	validationtestpublisher "github.com/kyma-project/eventing-tools/internal/validationtest/publisher"
)

// publisherCmd represents the publisher command
var publisherCmd = &cobra.Command{
	Use:   "publisher",
	Short: "Publish events compatible with kyma eventing",
	Run: func(cmd *cobra.Command, args []string) {
		switch loadtest {
		case true:
			loadtestpublisher.Start()
		case false:
			validationtestpublisher.Start()
		}
	},
}

func init() {
	rootCmd.AddCommand(publisherCmd)
}
