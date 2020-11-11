package float

/*

// Cmd .
var Cmd = &cobra.Command{
	Use:   "float [ID]",
	Short: "Deploys to the device/room/building with the given ID",
	Args:  args.ValidID,
	Run: func(cmd *cobra.Command, arg []string) {
		fmt.Printf("Deploying to %s\n", arg[0])
		fail := func(format string, a ...interface{}) {
			fmt.Printf(format, a...)
			os.Exit(1)
		}

		_, designation, err := args.GetDB()
		if err != nil {
			fail("error getting designation: %v", err)
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

		stream, err := client.Float(context.TODO(), &avcli.ID{Id: arg[0], Designation: designation})
		if err != nil {
			if s, ok := status.FromError(err); ok {
				switch s.Code() {
				case codes.Unavailable:
					fail("api is unavailable: %s\n", s.Err())
				default:
					fail("%s\n", s.Err())
				}
			}

			fail("unable to float: %s\n", err)
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
				fmt.Printf("there was an error floating to %s: %s\n", in.Id, in.Error)
			} else {
				fmt.Printf("Successfully floated to %s\n", in.Id)
			}
		}
	},
}
*/
