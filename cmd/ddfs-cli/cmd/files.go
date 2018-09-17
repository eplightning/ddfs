package cmd

import (
	"bytes"
	"context"
	"io"
	"os"

	"git.eplight.org/eplightning/ddfs/pkg/client"
	"git.eplight.org/eplightning/ddfs/pkg/monitor"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var fileCmd = &cobra.Command{
	Use:   "file",
	Short: "Manage files",
}

func init() {
	readCmd := &cobra.Command{
		Use:   "read [VOLUME]",
		Short: "Read file",
		Run:   readFile,
		Args:  cobra.ExactArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlag("start", cmd.Flags().Lookup("start"))
			viper.BindPFlag("end", cmd.Flags().Lookup("end"))
		},
	}
	readCmd.Flags().Int64("start", 0, "offset to start from")
	readCmd.Flags().Int64("end", 4096, "offset to end at (exclusive)")

	writeCmd := &cobra.Command{
		Use:   "write [VOLUME]",
		Short: "Write file",
		Run:   writeFile,
		Args:  cobra.ExactArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlag("start", cmd.Flags().Lookup("start"))
			viper.BindPFlag("end", cmd.Flags().Lookup("end"))
			viper.BindPFlag("file", cmd.Flags().Lookup("file"))
		},
	}
	writeCmd.Flags().Int64("start", 0, "offset to start from")
	writeCmd.Flags().Int64("end", 0, "offset to end at (exclusive), start + size if 0")
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
	combined := combinedClient(ctx, mon)

	err = combined.Read(ctx, args[0], viper.GetInt64("start"), viper.GetInt64("end"), os.Stdout)
	panicOnError(err)
}

func writeFile(cmd *cobra.Command, args []string) {
	mon, err := client.ConnectToMonitor(viper.GetStringSlice("monitorServers"))
	defer mon.Close()
	panicOnError(err)

	ctx := context.Background()
	combined := combinedClient(ctx, mon)

	start := viper.GetInt64("start")
	end := viper.GetInt64("end")

	var read io.Reader = os.Stdin
	file := viper.GetString("file")

	if file != "" && file != "-" {
		f, err := os.Open(file)
		panicOnError(err)
		read = f

		if end == 0 {
			finfo, err := f.Stat()
			panicOnError(err)
			end = start + finfo.Size()
		}
	} else {
		// read entire stdin if length is not user provided
		if end == 0 {
			data := new(bytes.Buffer)
			wr, err := io.Copy(data, os.Stdin)
			panicOnError(err)
			end = start + wr
			read = data
		}
	}

	err = combined.Write(ctx, args[0], start, end, io.LimitReader(read, end-start))
	panicOnError(err)
}

func combinedClient(ctx context.Context, mon monitor.Client) *client.CombinedClient {
	idx, err := client.NewIndexClient(ctx, mon)
	panicOnError(err)
	blk, err := client.NewBlockClient(ctx, mon)
	panicOnError(err)

	combined := client.NewCombinedClient(blk, idx)

	return combined
}
