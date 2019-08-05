package auth

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"runtime"

	"github.com/spf13/viper"
)

const port = ":7444"

type authCodeResponse struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
}

// GetWSO2Token .
func GetWSO2Token() (string, error) {
	if !viper.IsSet("wso2-key") || !viper.IsSet("wso2-secret") {
		// check env vars, set in the config file
		return "", fmt.Errorf("wso2 key/secret is not set")
	}

	if viper.IsSet("refresh-token") {
	} else {
		// get the refresh token
		code := getAuthCode(viper.GetString("wso2-key"))
		fmt.Printf("auth code: %s", code)
	}

	return "", nil
}

func getAuthCode(key string) string {
	codeChan := make(chan string)

	url := fmt.Sprintf("https://api.byu.edu/authorize?response_type=code&client_id=%s&redirect_uri=%s&scope=PRODUCTION", key, "http://localhost:7444")

	// run the server
	go func() {
		stop := make(chan struct{})
		srv := http.Server{
			Addr: port,
		}

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			code, ok := r.URL.Query()["code"]
			if !ok {
				log.Print("Auth Code not found")
			}

			io.WriteString(w, "success")
			codeChan <- code[len(code)-1]
			stop <- struct{}{}
		})

		fmt.Printf("Waiting for callback from wso2...\n")

		go srv.ListenAndServe()
		<-stop
		srv.Close()
	}()

	fmt.Printf("opening %s in the background\n", url)

	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
	}

	if err != nil {
		fmt.Printf("unable to open browser (%s). copy and paste the url into a browser", err)
	}

	code := <-codeChan
	return code
}
