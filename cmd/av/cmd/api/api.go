package api


/*

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
		var dockerDesignation string
		switch designation {
		case "dev":
			dockerDesignation = "development"
		case "stg":
			dockerDesignation = "stage"
		case "prd":
			dockerDesignation = "latest"
		}

		tempDockerHubTag := os.Getenv("DOCKER_HUB_TAG")
		tempSystemID := os.Getenv("SYSTEM_ID")

		os.Setenv("DOCKER_HUB_TAG", dockerDesignation)
		os.Setenv("SYSTEM_ID", args[0])
		c := exec.Command("docker-compose", "-f", fmt.Sprintf("%v/src/github.com/byuoitav/av-api/docker-compose-pull.yml", os.Getenv("GOPATH")), "pull")
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		err = c.Run()
		os.Setenv("DOCKER_HUB_TAG", tempDockerHubTag)
		os.Setenv("SYSTEM_ID", tempSystemID)
		if err != nil {
			fmt.Printf("Error running command: %v\n", err)
			os.Exit(1)
		}

		tempDockerHubTag = os.Getenv("DOCKER_HUB_TAG")
		tempSystemID = os.Getenv("SYSTEM_ID")
		tempDBAddress := os.Getenv("DB_ADDRESS")
		tempDBName := os.Getenv("DB_USERNAME")

		os.Setenv("DB_USERNAME", designation)
		finalAddr := strings.Replace(os.Getenv("DB_ADDRESS"), "dev", designation, 1)
		finalAddr = strings.Replace(finalAddr, "stg", designation, 1)
		finalAddr = strings.Replace(finalAddr, "prd", designation, 1)
		os.Setenv("DB_ADDRESS", finalAddr)

		os.Setenv("DOCKER_HUB_TAG", dockerDesignation)
		os.Setenv("SYSTEM_ID", args[0])
		c = exec.Command("docker-compose", "-f", fmt.Sprintf("%v/src/github.com/byuoitav/av-api/docker-compose-pull.yml", os.Getenv("GOPATH")), "up", "-d")
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		err = c.Run()
		os.Setenv("DOCKER_HUB_TAG", tempDockerHubTag)
		os.Setenv("SYSTEM_ID", tempSystemID)
		os.Setenv("DB_USERNAME", tempDBName)
		os.Setenv("DB_ADDRESS", tempDBAddress)
		if err != nil {
			fmt.Printf("Error running command: %v\n", err)
			os.Exit(1)
		}

	},
}

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "runs docker-compose down",
	Args:  args.Valid,
	Run: func(cmd *cobra.Command, args []string) {
		c := exec.Command("docker-compose", "-f", fmt.Sprintf("%v/src/github.com/byuoitav/av-api/docker-compose-pull.yml", os.Getenv("GOPATH")), "down")
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		err := c.Run()
		if err != nil {
			fmt.Printf("Error running command: %v\n", err)
			os.Exit(1)
		}
	},
}
*/
