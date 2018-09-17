package cmd

import (
	"os"
	"time"

	"git.eplight.org/eplightning/ddfs/pkg/api"
	"git.eplight.org/eplightning/ddfs/pkg/block"
	"google.golang.org/grpc"

	"git.eplight.org/eplightning/ddfs/pkg/util"
	"github.com/spf13/viper"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const BlockVersion = "0.1.0"

var rootCmd = &cobra.Command{
	Use:     "ddfs-block",
	Run:     runServer,
	Version: BlockVersion,
}

func init() {
	rootCmd.PersistentFlags().StringSlice("monitor-servers", []string{"localhost:7300"}, "monitor endpoints")
	rootCmd.PersistentFlags().String("listen", ":7301", "gRPC server listen address")
	rootCmd.PersistentFlags().String("data-path", "block-data", "where data should be stored")
	rootCmd.PersistentFlags().String("server-name", "", "server name")
	rootCmd.PersistentFlags().Int("meta-cache-count", 1024*1024, "number of metadata entries cached")
	rootCmd.PersistentFlags().Int("block-cache-count", 1024, "number of blocks cached")
	viper.BindPFlag("monitorServers", rootCmd.PersistentFlags().Lookup("monitor-servers"))
	viper.BindPFlag("listen", rootCmd.PersistentFlags().Lookup("listen"))
	viper.BindPFlag("dataPath", rootCmd.PersistentFlags().Lookup("data-path"))
	viper.BindPFlag("serverName", rootCmd.PersistentFlags().Lookup("server-name"))
	viper.BindPFlag("metaCacheCount", rootCmd.PersistentFlags().Lookup("meta-cache-count"))
	viper.BindPFlag("blockCacheCount", rootCmd.PersistentFlags().Lookup("block-cache-count"))
	viper.BindEnv("monitorServers", "MONITOR_SERVERS")
	viper.BindEnv("listen", "LISTEN")
	viper.BindEnv("dataPath", "DATA_PATH")
	viper.BindEnv("serverName", "SERVER_NAME")
	viper.BindEnv("metaCacheCount", "META_CACHE_COUNT")
	viper.BindEnv("blockCacheCount", "BLOCK_CACHE_COUNT")
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
	log.Info().Msgf("Starting ddfs-block %v on %v", BlockVersion, name)

	cc, err := grpc.Dial(viper.GetStringSlice("monitorServers")[0], grpc.WithTimeout(15*time.Second), grpc.WithBlock(), grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer cc.Close()
	log.Info().Msgf("Connected to monitor server %v", cc.Target())

	srv := util.NewGrpcServer(viper.GetString("listen"))
	manager := block.NewBlockManager(viper.GetString("dataPath"), viper.GetInt("metaCacheCount"), viper.GetInt("blockCacheCount"))

	util.InitSubsystems(srv, manager)

	service := block.NewBlockGrpc(manager)
	api.RegisterBlockStoreServer(srv.Server, service)

	util.StartSubsystems(srv, manager)
}
