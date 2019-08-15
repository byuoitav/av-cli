package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
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
	var delExt string
	switch runtime.GOOS {
	case "linux", "darwin":
		delExt = ".tmp"
	case "windows":
		delExt = ".old"
	}

	if len(delExt) > 0 {
		if bin, _ := os.Executable(); len(bin) > 0 {
			if _, err := os.Stat(bin + delExt); err == nil {
				if err = os.Remove(bin + delExt); err != nil {
					fmt.Printf("unable to remove extra version of %s: %s\n", color.GreenString("av"), err)
				} else {
					fmt.Printf("Removed an extra version of %s found on your computer.\n", color.GreenString("av"))
				}
			}
		}
	}

	if shouldTryUpdate() {
		err := update()
		if err != nil {
			fmt.Printf("unable to update: %s\n\n", err)
		}

		os.Exit(0) // so that the initial command doesn't run
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

	newBin, err := ioutil.TempFile("", "av")
	if err != nil {
		return fmt.Errorf("unable to create file for new version: %s", err)
	}
	defer newBin.Close()

	bar := pb.Full.Start64(asset.Size)
	barWriter := bar.NewProxyWriter(newBin)

	client := http.Client{}

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

	bin, err := os.Executable()
	if err != nil {
		return fmt.Errorf("unable to get current executable name: %s", err)
	}

	switch runtime.GOOS {
	case "linux", "darwin":
		err = move(newBin.Name(), bin)
		if err != nil {
			return fmt.Errorf("unable to replace old version with new version: %s", err)
		}

		err = os.Chmod(bin, 0755)
		if err != nil {
			return fmt.Errorf("unable to make new version executable. please run `chmod +x %s`", bin)
		}
	case "windows":
		// screw you windows
		err = os.Rename(bin, bin+".old")
		if err != nil {
			return fmt.Errorf("unable to move old version: %s", err)
		}

		cmd := []string{"cmd.exe", "/C", "start", "/b", "cmd.exe", "/C", "timeout", "/t", "1", "/nobreak", "&&", "start", "/b", "cmd.exe", "/C", "move", newBin.Name(), bin}

		err = exec.Command(cmd[0], cmd[1:]...).Run()
		if err != nil {
			return fmt.Errorf("unable to replace old version with new version: %s", err)
		}
	default:
		return fmt.Errorf("not sure how to update binary for %s. please replace %s with %s to finish the update", runtime.GOOS, bin, newBin.Name())
	}

	return nil
}

// makes sure that files can be moved across partitions (rename doesn't work in that case)
func move(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("unable to open src file: %s", err)
	}

	dstFile, err := os.Create(dst + ".tmp")
	if err != nil {
		return fmt.Errorf("unable to open dst file: %s", err)
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("unable to copy src->dst: %s", err)
	}

	err = os.Remove(src)
	if err != nil {
		return fmt.Errorf("unable to remove src file: %s", err)
	}

	err = os.Rename(dst+".tmp", dst)
	if err != nil {
		return fmt.Errorf("unable to rename copied file: %s", err)
	}

	return nil
}
