package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const CliVersion = "0.1.0"

var rootCmd = &cobra.Command{
	Use:     "ddfs-cli",
	Version: CliVersion,
}

func init() {
	rootCmd.PersistentFlags().StringSlice("monitor-servers", []string{"localhost:7300"}, "monitor endpoints")
	viper.BindPFlag("monitorServers", rootCmd.PersistentFlags().Lookup("monitor-servers"))
	viper.BindEnv("monitorServers", "MONITOR_SERVERS")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func printAsJson(value interface{}) {
	output, err := json.Marshal(value)
	panicOnError(err)

	fmt.Println(string(output))
}
