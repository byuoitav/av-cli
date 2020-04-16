package board

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/byuoitav/av-cli/cmd/args"
	"github.com/spf13/cobra"
)

// Cmd .
var Cmd = &cobra.Command{
	Use:   "board [device ID]",
	Short: "ssh into a pi",
	Long:  "open an ssh connection to a pi with the ability to enter commands",
	Args:  args.ValidDeviceID,
	Run: func(cmd *cobra.Command, args []string) {
		sshArgs := []string{"ssh", fmt.Sprintf("pi@%s.byu.edu", args[0])}
		sshPath, err := exec.LookPath("ssh")
		if err != nil {
			fmt.Printf("failure to find ssh executable: %v", err)
			os.Exit(1)
		}
		err = syscall.Exec(sshPath, sshArgs, os.Environ())
		if err != nil {
			fmt.Printf("board error: %v\n", err)
			os.Exit(1)
		}
	},
}
