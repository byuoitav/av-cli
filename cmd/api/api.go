package api

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/byuoitav/av-cli/cmd/args"
	arg "github.com/byuoitav/av-cli/cmd/args"
	"github.com/spf13/cobra"
)

func init() {
	Cmd.AddCommand(upCmd)
	Cmd.AddCommand(downCmd)
}

// Cmd .
var Cmd = &cobra.Command{
	Use:   "api",
	Short: "commands for managing the AV API",
}

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "runs docker-compose up",
	Args:  arg.ValidDeviceID,
	Run: func(cmd *cobra.Command, args []string) {
		_, designation, err := arg.GetDB()
		if err != nil {
			fmt.Printf("error getting designation: %v", err)
			os.Exit(1)
		}

		switch designation {
		case "dev":
			designation = "development"
		case "stg":
			designation = "stage"
		case "prd":
			designation = "latest"
		}

		tempDockerHubTag := os.Getenv("DOCKER_HUB_TAG")
		tempSystemID := os.Getenv("SYSTEM_ID")

		os.Setenv("DOCKER_HUB_TAG", designation)
		os.Setenv("SYSTEM_ID", args[0])
		out, err := exec.Command("docker-compose", "-f", fmt.Sprintf("%v/src/github.com/byuoitav/av-api/docker-compose-pull.yml", os.Getenv("GOPATH")), "pull").CombinedOutput()
		os.Setenv("DOCKER_HUB_TAG", tempDockerHubTag)
		os.Setenv("SYSTEM_ID", tempSystemID)
		if err != nil {
			fmt.Printf("Error running command: %v %s\n", err, out)
			os.Exit(1)
		}
		fmt.Printf("%s\n", out)

		tempDockerHubTag = os.Getenv("DOCKER_HUB_TAG")
		tempSystemID = os.Getenv("SYSTEM_ID")

		os.Setenv("DOCKER_HUB_TAG", designation)
		os.Setenv("SYSTEM_ID", args[0])
		out, err = exec.Command("docker-compose", "-f", fmt.Sprintf("%v/src/github.com/byuoitav/av-api/docker-compose-pull.yml", os.Getenv("GOPATH")), "up", "-d").CombinedOutput()
		os.Setenv("DOCKER_HUB_TAG", tempDockerHubTag)
		os.Setenv("SYSTEM_ID", tempSystemID)
		if err != nil {
			fmt.Printf("Error running command: %v %s\n", err, out)
			os.Exit(1)
		}
		fmt.Printf("%s\n", out)

	},
}

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "runs docker-compose down",
	Args:  args.Valid,
	Run: func(cmd *cobra.Command, args []string) {
		//tempDockerHubTag := os.Getenv("DOCKER_HUB_TAG")
		//tempSystemID := os.Getenv("SYSTEM_ID")

		//os.Setenv("DOCKER_HUB_TAG", designation)
		//os.Setenv("SYSTEM_ID", args[0])
		out, err := exec.Command("docker-compose", "-f", fmt.Sprintf("%v/src/github.com/byuoitav/av-api/docker-compose-pull.yml", os.Getenv("GOPATH")), "down").CombinedOutput()
		//os.Setenv("DOCKER_HUB_TAG", tempDockerHubTag)
		//os.Setenv("SYSTEM_ID", tempSystemID)
		if err != nil {
			fmt.Printf("Error running command: %v %s\n", err, out)
			os.Exit(1)
		}
		fmt.Printf("%s\n", out)
	},
}
