package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/byuoitav/av-cli/cmd/api"
	"github.com/byuoitav/av-cli/cmd/aws"
	"github.com/byuoitav/av-cli/cmd/board"
	"github.com/byuoitav/av-cli/cmd/db"
	"github.com/byuoitav/av-cli/cmd/float"
	"github.com/byuoitav/av-cli/cmd/pi"
	"github.com/byuoitav/av-cli/cmd/smee"
	"github.com/byuoitav/av-cli/cmd/swab"
	"github.com/fatih/color"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

func init() {
	// disable color if windows
	switch runtime.GOOS {
	case "windows":
		color.NoColor = true
	}

	cobra.OnInitialize(initConfig, initUpdate)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.av.yaml)")
	rootCmd.PersistentFlags().StringP("refresh-token", "t", "", "a wso2 refresh token to use")

	_ = viper.BindPFlag("wso2.refresh-token", rootCmd.PersistentFlags().Lookup("refresh-token"))

	viper.SetConfigType("yaml")
	viper.SetEnvPrefix("avcli")
	viper.AutomaticEnv()

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(float.Cmd)
	rootCmd.AddCommand(swab.Cmd)
	rootCmd.AddCommand(pi.Cmd)
	rootCmd.AddCommand(smee.Cmd)
	rootCmd.AddCommand(aws.Cmd)
	rootCmd.AddCommand(api.Cmd)
	rootCmd.AddCommand(db.Cmd)
	rootCmd.AddCommand(board.Cmd)
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
