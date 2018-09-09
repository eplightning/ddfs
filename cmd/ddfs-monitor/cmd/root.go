package cmd

import (
	"os"

	"git.eplight.org/eplightning/ddfs/pkg/api"
	"git.eplight.org/eplightning/ddfs/pkg/monitor"

	"git.eplight.org/eplightning/ddfs/pkg/util"
	"github.com/spf13/viper"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

const MonitorVersion = "0.1.0"

var rootCmd = &cobra.Command{
	Use:     "ddfs-monitor",
	Run:     runServer,
	Version: MonitorVersion,
}

func init() {
	rootCmd.PersistentFlags().StringSlice("etcd-servers", []string{"localhost:2379"}, "etcd endpoints")
	rootCmd.PersistentFlags().String("etcd-prefix", "ddfs/", "etcd prefix")
	rootCmd.PersistentFlags().String("listen", ":7300", "gRPC server listen address")
	viper.BindPFlag("etcdServers", rootCmd.PersistentFlags().Lookup("etcd-servers"))
	viper.BindPFlag("etcdPrefix", rootCmd.PersistentFlags().Lookup("etcd-prefix"))
	viper.BindPFlag("listen", rootCmd.PersistentFlags().Lookup("listen"))
	viper.BindEnv("etcdServers", "ETCD_SERVERS")
	viper.BindEnv("etcdPrefix", "ETCD_PREFIX")
	viper.BindEnv("listen", "LISTEN")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runServer(cmd *cobra.Command, args []string) {
	util.SetupLogging()
	log.Info().Msgf("Starting ddfs monitor %v", MonitorVersion)

	etcd := monitor.NewEtcdManager(viper.GetStringSlice("etcdServers"), viper.GetString("etcdPrefix"))
	srv := util.NewGrpcServer(viper.GetString("listen"))

	util.InitSubsystems(etcd, srv)

	service := monitor.NewMonitorGrpc(etcd)
	api.RegisterNodesServer(srv.Server, service)
	api.RegisterVolumesServer(srv.Server, service)
	api.RegisterSettingsServer(srv.Server, service)

	util.StartSubsystems(etcd, srv)
}
