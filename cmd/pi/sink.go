package pi

import (
	"fmt"
	"os"

	"github.com/byuoitav/av-cli/cmd/args"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

var sinkCmd = &cobra.Command{
	Use:   "sink [device ID]",
	Short: "reboot a pi",
	Long:  "ssh into a pi and reboot it",
	Args:  args.ValidDeviceID,
	Run: func(cmd *cobra.Command, args []string) {
		client, err := NewSSHClient("ITB-1101-CP1.byu.edu")
		if err != nil {
			fmt.Printf("unable to ssh into %s: %s", args[0], err)
			os.Exit(1)
		}
		defer client.Close()

		session, err := client.NewSession()
		if err != nil {
			fmt.Printf("unable to start new session: %s", err)
			client.Close()
			os.Exit(1)
		}

		fmt.Printf("Rebooting...\n")

		bytes, err := session.CombinedOutput("sudo reboot")
		if err != nil {
			switch err.(type) {
			case *ssh.ExitMissingError:
				fmt.Printf("Success.\n")
				return
			default:
				fmt.Printf("unable to reboot: %s\noutput on pi:\n%s\n", err, bytes)
				client.Close()
				os.Exit(1)
			}
		}

		fmt.Printf("unable to reboot:\n%s\n", bytes)
		os.Exit(1)
	},
}
