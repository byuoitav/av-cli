package smee

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/byuoitav/av-cli/cli/cmd/args"
	"github.com/byuoitav/av-cli/cli/cmd/wso2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	closeIssueCmd.Flags().StringP("access-key", "k", "", "the access key to use to authenticate against smee")
	_ = viper.BindPFlag("smee.access-key", closeIssueCmd.Flags().Lookup("access-key"))
}

var closeIssueCmd = &cobra.Command{
	Use:   "closeIssue [room ID]",
	Short: "Close a room issue in smee",
	Args:  args.ValidRoomID,
	Run: func(cmd *cobra.Command, args []string) {
		// get key from viper/flags. give preference to flag
		if !viper.IsSet("smee.access-key") {
			fmt.Printf("smee access-key not set. include the access key with -k [key] or add it to your config.\n")
			os.Exit(1)
		}
		key := viper.GetString("smee.access-key")

		id, err := wso2.GetIDInfo()
		if err != nil {
			fmt.Printf("unable to get id info: %s\n", err)
			os.Exit(1)
		}

		url := fmt.Sprintf("https://smee.avs.byu.edu/issues/%s/resolve", args[0])

		body, err := json.Marshal(map[string]interface{}{
			"resolution-code": "Manual Removal",
			"notes":           fmt.Sprintf("%s manually removed room issue through av-cli", id.NetID),
		})
		if err != nil {
			fmt.Printf("unable to build marshal request body: %s\n", err)
			os.Exit(1)
		}

		req, err := http.NewRequest("PUT", url, bytes.NewReader(body))
		if err != nil {
			fmt.Printf("unable to build request: %s\n", err)
			os.Exit(1)
		}

		req.Header.Add("content-type", "application/json")
		req.Header.Add("x-av-access-key", key)
		req.Header.Add("x-av-user", id.NetID)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("unable to make request: %s\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		if resp.StatusCode/100 != 2 {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("unable to close issue; response code %v. unable to read response body: %s\n", resp.StatusCode, err)
				os.Exit(1)
			}

			fmt.Printf("unable to close issue: %s\n", body)
			os.Exit(1)
		}

		fmt.Printf("Room issue closed.\n")
	},
}
