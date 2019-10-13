package cmd

import (
	"context"
	"time"

	"github.com/eplightning/ddfs/pkg/api"
	"github.com/eplightning/ddfs/pkg/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var volumeCmd = &cobra.Command{
	Use:   "volume",
	Short: "Manage volumess",
}

func init() {
	watchCmd := &cobra.Command{
		Use:   "watch",
		Short: "Watch volumes",
		Run:   watchVolumes,
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlag("revision", cmd.Flags().Lookup("revision"))
		},
	}
	watchCmd.Flags().Int64("revision", 1, "revision to start from")

	modifyCmd := &cobra.Command{
		Use:   "modify [VOLUME]",
		Short: "Modify volumes",
		Run:   modifyVolumes,
		Args:  cobra.ExactArgs(1),
	}
	modifyCmd.Flags().Int64("shards", 1, "how many shards")
	modifyCmd.MarkFlagRequired("shards")

	volumeCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List volumes",
		Run:   listVolumes,
	})
	volumeCmd.AddCommand(watchCmd)
	volumeCmd.AddCommand(modifyCmd)

	rootCmd.AddCommand(volumeCmd)
}

func listVolumes(cmd *cobra.Command, args []string) {
	mon, err := client.ConnectToMonitor(viper.GetStringSlice("monitorServers"))
	defer mon.Close()
	panicOnError(err)

	context, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	resp, err := mon.Get(context, &api.GetVolumesRequest{})
	panicOnError(err)
	printAsJson(resp)
}

func watchVolumes(cmd *cobra.Command, args []string) {
	mon, err := client.ConnectToMonitor(viper.GetStringSlice("monitorServers"))
	defer mon.Close()
	panicOnError(err)

	context := context.TODO()
	watch, err := client.WatchVolumes(mon, context, viper.GetInt64("revision"))
	panicOnError(err)

	for v := range watch {
		printAsJson(v)
	}
}

func modifyVolumes(cmd *cobra.Command, args []string) {
	mon, err := client.ConnectToMonitor(viper.GetStringSlice("monitorServers"))
	defer mon.Close()
	panicOnError(err)
	shards, err := cmd.Flags().GetInt64("shards")
	panicOnError(err)

	context, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	resp, err := mon.Modify(context, &api.ModifyVolumeRequest{
		Name:   args[0],
		Shards: shards,
	})
	panicOnError(err)
	printAsJson(resp)
}
