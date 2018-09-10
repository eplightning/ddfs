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
	rootCmd.PersistentFlags().String("listen", ":7301", "gRPC server listen address")
	rootCmd.PersistentFlags().String("data-path", "index-data", "where data should be stored")
	viper.BindPFlag("monitorServers", rootCmd.PersistentFlags().Lookup("monitor-servers"))
	viper.BindPFlag("listen", rootCmd.PersistentFlags().Lookup("listen"))
	viper.BindPFlag("dataPath", rootCmd.PersistentFlags().Lookup("data-path"))
	viper.BindEnv("monitorServers", "MONITOR_SERVERS")
	viper.BindEnv("listen", "LISTEN")
	viper.BindEnv("dataPath", "DATA_PATH")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runServer(cmd *cobra.Command, args []string) {
	util.SetupLogging()
	log.Info().Msgf("Starting ddfs index %v", IndexVersion)

	cc, err := grpc.Dial(viper.GetStringSlice("monitorServers")[0], grpc.WithTimeout(15*time.Second), grpc.WithBlock(), grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer cc.Close()
	mon := monitor.FromClientConn(cc)
	log.Info().Msg("Connected to monitor server " + cc.Target())

	shards := index.NewShardManager(mon, 10*1024, viper.GetString("dataPath"))
	srv := util.NewGrpcServer(viper.GetString("listen"))

	util.InitSubsystems(shards, srv)

	service := index.NewIndexGrpc(shards)
	api.RegisterIndexStoreServer(srv.Server, service)

	util.StartSubsystems(shards, srv)
}
