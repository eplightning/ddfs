package cmd

import (
	"context"
	"io"
	"os"

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

	writeCmd := &cobra.Command{
		Use:        "write",
		Short:      "Write file",
		Run:        writeFile,
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"volume"},
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlag("start", cmd.Flags().Lookup("start"))
			viper.BindPFlag("end", cmd.Flags().Lookup("end"))
			viper.BindPFlag("file", cmd.Flags().Lookup("file"))
		},
	}
	writeCmd.Flags().Int64("start", 0, "offset to start from")
	writeCmd.Flags().Int64("end", 4096, "offset to end at (exclusive)")
	writeCmd.Flags().StringP("file", "f", "-", "file to write")

	fileCmd.AddCommand(readCmd)
	fileCmd.AddCommand(writeCmd)

	rootCmd.AddCommand(fileCmd)
}

func readFile(cmd *cobra.Command, args []string) {
	mon, err := client.ConnectToMonitor(viper.GetStringSlice("monitorServers"))
	defer mon.Close()
	panicOnError(err)

	ctx := context.Background()
	idx, err := client.NewIndexClient(ctx, mon)
	panicOnError(err)
	blk, err := client.NewBlockClient(ctx, mon)
	panicOnError(err)
	combined := client.NewCombinedClient(blk, idx)
	err = combined.Read(ctx, args[0], viper.GetInt64("start"), viper.GetInt64("end"), os.Stdout)
	panicOnError(err)
}

func writeFile(cmd *cobra.Command, args []string) {
	mon, err := client.ConnectToMonitor(viper.GetStringSlice("monitorServers"))
	defer mon.Close()
	panicOnError(err)

	ctx := context.Background()
	idx, err := client.NewIndexClient(ctx, mon)
	panicOnError(err)
	blk, err := client.NewBlockClient(ctx, mon)
	panicOnError(err)
	combined := client.NewCombinedClient(blk, idx)

	var read io.Reader = os.Stdin
	file := viper.GetString("file")
	if file != "" && file != "-" {
		read, err = os.Open(file)
		panicOnError(err)
	}

	start := viper.GetInt64("start")
	end := viper.GetInt64("end")

	err = combined.Write(ctx, args[0], start, end, io.LimitReader(read, end-start))
	panicOnError(err)
}
