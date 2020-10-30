package pi

/*

var sinkCmd = &cobra.Command{
	Use:   "sink [device ID]",
	Short: "reboot a pi",
	Long:  "ssh into a pi and reboot it",
	Args:  args.ValidDeviceID,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Rebooting %s\n", args[0])
		fail := func(format string, a ...interface{}) {
			fmt.Printf(format, a...)
			os.Exit(1)
		}

		idToken := wso2.GetIDToken()

		auth := avcli.Auth{
			Token: idToken,
			User:  "",
		}

		client, err := avcli.NewClient(viper.GetString("api"), auth)
		if err != nil {
			fail("unable to create client: %v\n", err)
		}

		stream, err := client.Sink(context.TODO(), &avcli.ID{Id: args[0]})
		if err != nil {
			if s, ok := status.FromError(err); ok {
				switch s.Code() {
				case codes.Unavailable:
					fail("api is unavailable: %s\n", s.Err())
				default:
					fail("%s\n", s.Err())
				}
			}

			fail("unable to reboot: %v\n", err)
		}

		for {
			in, err := stream.Recv()
			switch {
			case errors.Is(err, io.EOF):
				return
			case err != nil:
				fmt.Printf("error: %s\n", err)
				return
			}

			if in.Error != "" {
				fmt.Printf("there was an error rebooting %s: %s\n", in.Id, in.Error)
			} else {
				fmt.Printf("Successfully rebooted %s\n", in.Id)
			}
		}
	},
}
*/
