package cmd

import (
	"github.com/spf13/cobra"

	loadtestsubscriber "github.com/kyma-project/eventing-tools/internal/loadtest/subscriber"
)

var subscriberPort int

// subscriberCmd represents the subscriber command
var subscriberCmd = &cobra.Command{
	Use:   "subscriber",
	Short: "Listen on a given port for cloudevents",
	Run: func(cmd *cobra.Command, args []string) {
		loadtestsubscriber.Start(subscriberPort)
	},
}

func init() {
	rootCmd.AddCommand(subscriberCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// subscriberCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// subscriberCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	subscriberCmd.Flags().IntVarP(&subscriberPort, "listen-port", "p", 8080, "listen on port for incoming events")

}
