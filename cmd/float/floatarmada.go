package float

/*
const (
	dev  = "development"
	stg  = "stage"
	prd  = "production"
	test = "testing"
)

var armadaCmd = &cobra.Command{
	Use:   "armada [designation ID]",
	Short: "Deploys to all rooms with the given designation",
	Args: func(cmd *cobra.Command, args []string) error {
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		db, designation, err := arg.GetDB()
		if err != nil {
			fmt.Printf("Couldn't get the database: %v", err)
			return
		}

		fmt.Printf("Deploying to all %s rooms\n", designation)

		err = floatarmada(db, designation)
		if err != nil {
			fmt.Printf("Error floating armada: %v\n", err)
			return
		}

	},
}

func floatarmada(db db.DB, designation string) error {
	rooms, err := db.GetRoomsByDesignation(designation)
	if err != nil {
		return fmt.Errorf("unable to get rooms from database: %s", err)
	}

	if len(rooms) == 0 {
		return fmt.Errorf("no %s rooms found", designation)
	}

	var bars []*pb.ProgressBar
	for _, room := range rooms {
		devs, err := db.GetDevicesByRoom(room.ID)
		if err != nil {
			return fmt.Errorf("couldn't get devices for room %v: %v", room.ID, err)
		}
		bar := pb.New(len(devs) + 6).SetWidth(50).Format(fmt.Sprintf("%s [\x00=\x00>\x00-\x00]", room.ID))
		bar.ShowCounters = false
		bars = append(bars, bar)

	}

	wg := sync.WaitGroup{}

	failedCount := 0
	failedList := ""
	pool := pb.NewPool(bars...)
	pool.Start()

	for i := range rooms {
		wg.Add(1)

		go func(idx int) {
			defer wg.Done()
			fmt.Printf("Deploying to %s\n", rooms[idx].ID)
			err := floatsquadronWithBar(db, rooms[idx].ID, designation, bars[idx])
			if err != nil {
				failedList = fmt.Sprintf("%v%v: %v\n", failedList, rooms[idx].ID, err)
				failedCount++
				bars[idx].Finish()
				return
			}
		}(i)
	}
	wg.Wait()
	pool.Stop()

	fmt.Printf("%v failures:\n%v\n", failedCount, failedList)

	return nil
}
*/
