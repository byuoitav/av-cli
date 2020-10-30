package pi

/*
var fixTimeCmd = &cobra.Command{
	Use:   "fixtime [device ID]",
	Short: "fix a pi who's time is off",
	Long:  "force an NTP sync of a pi to fix time drift",
	Args:  args.ValidDeviceID,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Fixing time on %s\n", args[0])
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

		stream, err := client.FixTime(context.TODO(), &avcli.ID{Id: args[0]}, grpc.PerRPCCredentials(auth))
		if err != nil {
			if s, ok := status.FromError(err); ok {
				switch s.Code() {
				case codes.Unavailable:
					fail("api is unavailable: %s\n", s.Err())
				default:
					fail("%s\n", s.Err())
				}
			}

			fail("unable to fix time: %v\n", err)
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
				fmt.Printf("there was an error fixing time on %s: %s\n", in.Id, in.Error)
			} else {
				fmt.Printf("Time fixed on %s\n", in.Id)
			}
		}
	},
}
*/
