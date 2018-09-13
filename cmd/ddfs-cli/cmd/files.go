package cmd

import (
	"context"

	"git.eplight.org/eplightning/ddfs/pkg/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var fileCmd = &cobra.Command{
	Use:   "file",
	Short: "Manage files",
}

func init() {
	readCmd := &cobra.Command{
		Use:        "read",
		Short:      "Read file",
		Run:        readFile,
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"volume"},
		PreRun: func(cmd *cobra.Command, args []string) {
			// viper.BindPFlag("revision", cmd.Flags().Lookup("revision"))
		},
	}
	// readCmd.Flags().Int64("revision", 1, "revision to start from")

	fileCmd.AddCommand(readCmd)

	rootCmd.AddCommand(fileCmd)
}

func readFile(cmd *cobra.Command, args []string) {
	mon, err := client.ConnectToMonitor(viper.GetStringSlice("monitorServers"))
	defer mon.Close()
	panicOnError(err)

	ctx := context.Background()
	idx, err := client.NewIndexClient(ctx, mon)
	panicOnError(err)
	idx.GetRange(ctx, args[0], 0, 100)
}
