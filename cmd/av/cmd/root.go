package cmd

import (
	"fmt"
	"os"
	"runtime"

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
	rootCmd.PersistentFlags().StringP("api", "a", "", "url of the av-cli API")
	rootCmd.PersistentFlags().StringP("refresh-token", "t", "", "a wso2 refresh token to use")

	_ = viper.BindPFlag("wso2.refresh-token", rootCmd.PersistentFlags().Lookup("refresh-token"))
	_ = viper.BindPFlag("api", rootCmd.PersistentFlags().Lookup("api"))

	viper.SetConfigType("yaml")
	viper.SetEnvPrefix("avcli")
	viper.AutomaticEnv()

	//rootCmd.AddCommand(versionCmd)
	//rootCmd.AddCommand(updateCmd)
	//rootCmd.AddCommand(float.Cmd)
	//rootCmd.AddCommand(swab.Cmd)
	//rootCmd.AddCommand(pi.Cmd)
	//rootCmd.AddCommand(smee.Cmd)
	//rootCmd.AddCommand(aws.Cmd)
	//rootCmd.AddCommand(api.Cmd)
	//rootCmd.AddCommand(db.Cmd)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Printf("unable to initialize config: %s", err)
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

	// set default values
	if len(viper.GetString("api")) > 0 {
		if err := viper.WriteConfig(); err != nil {
			fmt.Printf("unable to save values to config: %s\n", err)
			os.Exit(1)
		}
	}

	viper.SetDefault("api", "cli.av.byu.edu:443")
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
