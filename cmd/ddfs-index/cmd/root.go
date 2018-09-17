package cmd

import (
	"os"
	"time"

	"git.eplight.org/eplightning/ddfs/pkg/api"
	"git.eplight.org/eplightning/ddfs/pkg/index"
	"git.eplight.org/eplightning/ddfs/pkg/monitor"
	"google.golang.org/grpc"

	"git.eplight.org/eplightning/ddfs/pkg/util"
	"github.com/spf13/viper"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const IndexVersion = "0.1.0"

var rootCmd = &cobra.Command{
	Use:     "ddfs-index",
	Run:     runServer,
	Version: IndexVersion,
}

func init() {
	rootCmd.PersistentFlags().StringSlice("monitor-servers", []string{"localhost:7300"}, "monitor endpoints")
	rootCmd.PersistentFlags().String("listen", ":7302", "gRPC server listen address")
	rootCmd.PersistentFlags().String("data-path", "index-data", "where data should be stored")
	rootCmd.PersistentFlags().String("server-name", "", "server name")
	rootCmd.PersistentFlags().Int("slice-cache-count", 10*1024, "number of shard slices cached")
	rootCmd.PersistentFlags().Int32("entries-per-slice", 1024, "number of entries per shard slice")
	viper.BindPFlag("monitorServers", rootCmd.PersistentFlags().Lookup("monitor-servers"))
	viper.BindPFlag("listen", rootCmd.PersistentFlags().Lookup("listen"))
	viper.BindPFlag("dataPath", rootCmd.PersistentFlags().Lookup("data-path"))
	viper.BindPFlag("serverName", rootCmd.PersistentFlags().Lookup("server-name"))
	viper.BindPFlag("sliceCacheCount", rootCmd.PersistentFlags().Lookup("slice-cache-count"))
	viper.BindPFlag("entriesPerSlice", rootCmd.PersistentFlags().Lookup("entries-per-slice"))
	viper.BindEnv("monitorServers", "MONITOR_SERVERS")
	viper.BindEnv("listen", "LISTEN")
	viper.BindEnv("dataPath", "DATA_PATH")
	viper.BindEnv("serverName", "SERVER_NAME")
	viper.BindEnv("sliceCacheCount", "SLICE_CACHE_COUNT")
	viper.BindEnv("entriesPerSlice", "ENTRIES_PER_SLICE")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runServer(cmd *cobra.Command, args []string) {
	var err error
	name := viper.GetString("serverName")

	if name == "" {
		name, err = os.Hostname()
		if err != nil {
			panic(err)
		}
		if name == "" {
			panic("Empty server name")
		}
	}

	util.SetupLogging()
	log.Info().Msgf("Starting ddfs-index %v on %v", IndexVersion, name)

	cc, err := grpc.Dial(viper.GetStringSlice("monitorServers")[0], grpc.WithTimeout(15*time.Second), grpc.WithBlock(), grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer cc.Close()
	mon := monitor.FromClientConn(cc)
	log.Info().Msgf("Connected to monitor server %v", cc.Target())

	shards := index.NewShardManager(
		mon, viper.GetInt("sliceCacheCount"), viper.GetInt32("entries-per-slice"), viper.GetString("dataPath"), viper.GetString("serverName"))
	srv := util.NewGrpcServer(viper.GetString("listen"))

	util.InitSubsystems(shards, srv)

	service := index.NewIndexGrpc(shards)
	api.RegisterIndexStoreServer(srv.Server, service)

	util.StartSubsystems(shards, srv)
}
