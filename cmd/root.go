package cmd

import (
	"fmt"
	"os"

	flightdeck "github.com/byuoitav/av-cli/cmd/flight-deck"
	"github.com/byuoitav/av-cli/cmd/pi"
	"github.com/spf13/cobra"
)

func init() {
	// add all subcommands here
	rootCmd.AddCommand(flightdeck.FloatshipCmd)
	rootCmd.AddCommand(pi.SwabCmd)
	rootCmd.AddCommand(pi.SwabRoomCmd)
}

var rootCmd = &cobra.Command{
	Use:   "av",
	Short: "BYU OIT AV's cli",
	Long:  "",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		/*
			// TODO check if already logged in

			// prompt for username/password
			fmt.Printf("Logging into CAS...\n")

			unamePrompt := promptui.Prompt{
				Label: "Username",
			}

			passPrompt := promptui.Prompt{
				Label: "Password",
				Mask:  '*',
			}

			uname, err := unamePrompt.Run()
			if err != nil {
				fmt.Printf("prompt failed: %v\n", err)
				os.Exit(1)
			}

			pass, err := passPrompt.Run()
			if err != nil {
				fmt.Printf("prompt failed: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("uname: %s, pass: %s\n", uname, pass)

			// authenticate with cas
			cas.Login(uname, pass)

			os.Exit(1)
		*/
	},
}

// Execute .
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
