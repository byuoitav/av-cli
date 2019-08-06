package cmd

import (
	"fmt"
	"os"

	"github.com/byuoitav/av-cli/cmd/pi"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.av.yaml)")
	rootCmd.PersistentFlags().StringP("refresh-token", "t", "", "a wso2 refresh token to use")

	viper.BindPFlag("wso2.refresh-token", rootCmd.PersistentFlags().Lookup("refresh-token"))

	viper.SetConfigType("yaml")
	viper.SetEnvPrefix("avcli")
	viper.AutomaticEnv()

	// add all subcommands here
	rootCmd.AddCommand(floatCmd)

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
			fmt.Printf("unable to initalize config: %s", err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.SetConfigName(".av")
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if cfgFile != "" {
				// if they gave us an invalid config location
				fmt.Printf("no config file found at %s\n", cfgFile)
				os.Exit(1)
			}

			home, err := homedir.Dir()
			if err != nil {
				fmt.Printf("unable to create default config file: %s\n", err)
				os.Exit(1)
			}

			if err = viper.WriteConfigAs(home + "/.av.yaml"); err != nil {
				fmt.Printf("unable to create default config file: %s\n", err)
				os.Exit(1)
			}

			fmt.Printf("created empty config file\n")
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
