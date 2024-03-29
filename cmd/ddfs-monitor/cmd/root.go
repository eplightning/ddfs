package cmd

import (
	"os"

	"github.com/eplightning/ddfs/pkg/api"
	"github.com/eplightning/ddfs/pkg/monitor"

	"github.com/eplightning/ddfs/pkg/util"
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
	rootCmd.PersistentFlags().String("bootstrap-file", "", "Config bootstrap data")
	rootCmd.PersistentFlags().Bool("bootstrap-force", false, "Force cluster remake")
	viper.BindPFlag("etcdServers", rootCmd.PersistentFlags().Lookup("etcd-servers"))
	viper.BindPFlag("etcdPrefix", rootCmd.PersistentFlags().Lookup("etcd-prefix"))
	viper.BindPFlag("listen", rootCmd.PersistentFlags().Lookup("listen"))
	viper.BindPFlag("bootstrapFile", rootCmd.PersistentFlags().Lookup("bootstrap-file"))
	viper.BindPFlag("bootstrapForce", rootCmd.PersistentFlags().Lookup("bootstrap-force"))
	viper.BindEnv("etcdServers", "ETCD_SERVERS")
	viper.BindEnv("etcdPrefix", "ETCD_PREFIX")
	viper.BindEnv("listen", "LISTEN")
	viper.BindEnv("bootstrapFile", "BOOTSTRAP_FILE")
	viper.BindEnv("bootstrapForce", "BOOTSTRAP_FORCE")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runServer(cmd *cobra.Command, args []string) {
	util.SetupLogging()
	log.Info().Msgf("Starting ddfs monitor %v", MonitorVersion)

	etcd := monitor.NewEtcdManager(
		viper.GetStringSlice("etcdServers"), viper.GetString("etcdPrefix"), viper.GetString("bootstrapFile"), viper.GetBool("bootstrapForce"))
	srv := util.NewGrpcServer(viper.GetString("listen"))

	util.InitSubsystems(etcd, srv)

	service := monitor.NewMonitorGrpc(etcd)
	api.RegisterNodesServer(srv.Server, service)
	api.RegisterVolumesServer(srv.Server, service)
	api.RegisterSettingsServer(srv.Server, service)

	util.StartSubsystems(etcd, srv)
}
