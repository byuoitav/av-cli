package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func initUpdate() {
	if shouldTryUpdate() {
		err := update()
		if err != nil {
			fmt.Printf("unable to update: %s\n\n", err)
		}
	}
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Updates av-cli to the newest version",
	Run: func(cmd *cobra.Command, args []string) {
		update()
	},
}

type release struct {
	ID          int       `json:"id"`
	URL         string    `json:"url"`
	Tag         string    `json:"tag_name"`
	Target      string    `json:"target"`
	Name        string    `json:"name"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
	CreatedAt   time.Time `json:"created_at"`
	PublishedAt time.Time `json:"published_at"`
	Body        string    `json:"body"`

	Author struct {
		Login string `json:"login"`
	} `json:"author"`

	Assets []asset `json:"assets"`
}

type asset struct {
	URL         string    `json:"url"`
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	ContentType string    `json:"content_type"`
	State       string    `json:"state"`
	CreatedAt   time.Time `json:"created_at"`
	PublishedAt time.Time `json:"published_at"`
	Size        int64     `json:"size"`
}

func shouldTryUpdate() bool {
	if len(version) == 0 {
		return false
	}

	if time.Now().Sub(viper.GetTime("last-update-check")).Hours() > 24 {
		return true
	}

	return false
}

func update() error {
	fmt.Printf("Checking for an update for %s...\n", color.GreenString("av"))

	release, asset, err := getUpdateAsset()
	if err != nil {
		return fmt.Errorf("unable to get update asset: %s", err)
	}

	defer func() {
		viper.Set("last-update-check", time.Now())
		viper.WriteConfig()
	}()

	if release.Tag == version {
		fmt.Printf("Already have latest version.\n")
		return nil
	}

	if len(version) > 0 {
		fmt.Printf("A newer version (%s) is available - you have %s. Changelog:\n", release.Tag, version)
	} else {
		fmt.Printf("A newer version (%s) is available - you have %s.\n", release.Tag, color.New(color.Bold, color.Underline).Sprint("an unversioned build. Be careful updating"))
		path, err := os.Executable()
		if err == nil {
			fmt.Printf("%s will be overwritten on your computer if you choose to update.\n", path)
		}

		fmt.Printf("Changelog:\n")
	}

	fmt.Printf("\n%s\n%s\n\n", color.New(color.Underline, color.FgCyan).Sprint(release.Name), release.Body)

	prompt := promptui.Select{
		Label: "Would you like to update right now?",
		Items: []string{"Yes", "No"},
	}

	_, result, err := prompt.Run()
	if err != nil {
		return err
	}

	if !strings.EqualFold(result, "Yes") {
		fmt.Printf("\n")
		return nil
	}

	err = updateSelf(asset)
	if err != nil {
		fmt.Printf("unable to update: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully updated.\n")
	return nil
}

func getUpdateAsset() (release, asset, error) {
	req, err := http.NewRequest(http.MethodGet, "https://api.github.com/repos/byuoitav/av-cli/releases/latest", nil)
	if err != nil {
		return release{}, asset{}, fmt.Errorf("unable to build http request: %s", err)
	}

	client := http.Client{
		Timeout: 4 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return release{}, asset{}, fmt.Errorf("unable to send request to api.github.com: %s", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return release{}, asset{}, fmt.Errorf("unable to read response from api.github.com: %s", err)
	}

	if resp.StatusCode/2 != 100 {
		return release{}, asset{}, fmt.Errorf("%v response code recieved from api.github.com. response body: %s", resp.StatusCode, body)
	}

	var rel release
	err = json.Unmarshal(body, &rel)
	if err != nil {
		return release{}, asset{}, fmt.Errorf("unable to parse response from api.github.com: %s. response body: %s", err, body)
	}

	if rel.Tag == version {
		// no update required
		return rel, asset{}, nil
	}

	// find the correct binary
	expectedName := fmt.Sprintf("av-%s-%s", runtime.GOOS, runtime.GOARCH)

	for i := range rel.Assets {
		if strings.EqualFold(expectedName, rel.Assets[i].Name) {
			return rel, rel.Assets[i], nil
		}
	}

	return rel, asset{}, fmt.Errorf("newer version availabe, but no binary matching '%s' was found", expectedName)
}

func updateSelf(asset asset) error {
	req, err := http.NewRequest(http.MethodGet, asset.URL, nil)
	if err != nil {
		return fmt.Errorf("unable to create request: %s", err)
	}

	req.Header.Add("Accept", "application/octet-stream")

	client := http.Client{
		Timeout: 30 * time.Second,
	}

	fileName := fmt.Sprintf("av-%v", asset.ID)
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		return fmt.Errorf("unable to open file to open: %s", err)
	}
	defer file.Close()

	bar := pb.Full.Start64(asset.Size)
	barWriter := bar.NewProxyWriter(file)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("unable to send request: %s", err)
	}
	defer resp.Body.Close()

	n, err := io.Copy(barWriter, resp.Body)
	switch {
	case err != nil:
		return fmt.Errorf("unable to save file: %s", err)
	case n != asset.Size:
		return fmt.Errorf("unable to save file: wrote %v/%v bytes", n, asset.Size)
	}

	bar.Finish()

	binName, err := os.Executable()
	if err != nil {
		return fmt.Errorf("unable to get current executable name: %s", err)
	}

	err = os.Rename(fileName, binName)
	if err != nil {
		return fmt.Errorf("unable to get move new version to correct path: %s", err)
	}

	return nil
}
