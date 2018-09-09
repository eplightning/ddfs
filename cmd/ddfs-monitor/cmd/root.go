package cmd

import (
	"os"

	"git.eplight.org/eplightning/ddfs/pkg/util"
	"github.com/spf13/viper"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "monitor",
	Run: runServer,
}

func init() {
	rootCmd.PersistentFlags().StringSlice("etcd-servers", []string{"localhost:2379"}, "etcd endpoints")
	rootCmd.PersistentFlags().String("etcd-prefix", "ddfs", "etcd prefix")
	viper.BindPFlag("etcdServers", rootCmd.PersistentFlags().Lookup("etcd-servers"))
	viper.BindPFlag("etcdPrefix", rootCmd.PersistentFlags().Lookup("etcd-prefix"))
	viper.BindEnv("etcdServers", "ETCD_SERVERS")
	viper.BindEnv("etcdPrefix", "ETCD_PREFIX")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runServer(cmd *cobra.Command, args []string) {
	util.SetupLogging()

	log.Info().Msgf("Hello world %v", viper.GetStringSlice("etcdServers"))

	stopCh, _ := util.SetupSignalHandler()
	<-stopCh
}
