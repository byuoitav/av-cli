package pi

import (
	"fmt"
	"os"
	"strings"

	"github.com/byuoitav/av-cli/cli/cmd/args"
	"github.com/spf13/cobra"
)

var fixTimeCmd = &cobra.Command{
	Use:   "fixtime [device ID]",
	Short: "fix a pi who's time is off",
	Long:  "force an NTP sync of a pi to fix time drift",
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

		fmt.Printf("Fixing time on pi...\n")

		bytes, err := session.CombinedOutput("date; sudo ntpdate tick.byu.edu && date")
		if err != nil {
			fmt.Printf("unable to run fix time command: %s\noutput on pi:\n%s\n", err, bytes)
			client.Close()
			os.Exit(1)
		}

		f := func(c rune) bool {
			return c == 0x0a
		}

		split := strings.FieldsFunc(string(bytes), f)
		if len(split) != 3 {
			fmt.Printf("Weird response while updating time:\n%s\n", bytes)
			client.Close()
			os.Exit(1)
		}

		fmt.Printf("Successfully updated time to: %s\n", split[2])
	},
}
