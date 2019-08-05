package cmd

import (
	"fmt"
	"os"

	flightdeck "github.com/byuoitav/av-cli/cmd/flight-deck"
	"github.com/byuoitav/av-cli/cmd/pi"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.av.yaml)")

	// add all subcommands here

	//Deploying
	rootCmd.AddCommand(flightdeck.FloatshipCmd)
	rootCmd.AddCommand(flightdeck.FloatfleetCmd)
	rootCmd.AddCommand(flightdeck.FloatarmadaCmd)

	//Swab
	rootCmd.AddCommand(pi.SwabCmd)
	rootCmd.AddCommand(pi.SwabRoomCmd)
	rootCmd.AddCommand(pi.SwabBuildingCmd)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.SetConfigName(".av")
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Printf("no config file found\n")
		} else {
			fmt.Printf("unable to read config: %s\n", err)
			os.Exit(1)
		}
	}
}

var rootCmd = &cobra.Command{
	Use:   "av",
	Short: "BYU OIT AV's cli",
}

// Execute .
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
