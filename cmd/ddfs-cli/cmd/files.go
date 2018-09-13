package cmd

import (
	"context"
	"fmt"

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
			viper.BindPFlag("start", cmd.Flags().Lookup("start"))
			viper.BindPFlag("end", cmd.Flags().Lookup("end"))
		},
	}
	readCmd.Flags().Int64("start", 0, "offset to start from")
	readCmd.Flags().Int64("end", 4096, "offset to end at (exclusive)")

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
	ranges, err := idx.GetRange(ctx, args[0], viper.GetInt64("start"), viper.GetInt64("end"))
	for _, r := range ranges {
		fmt.Printf("%v", r)
	}
	fmt.Printf("%v %v", ranges, err)
}
