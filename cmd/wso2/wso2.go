package wso2

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/spf13/viper"
)

var (
	accessToken string
	once        sync.Once
)

type config struct {
	clientID     string
	clientSecret string
	redirect     string
	port         int
}

type authCodeResponse struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token"`
}

// GetToken .
func GetToken() string {
	once.Do(func() {
		var err error
		accessToken, err = getToken()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	})

	return accessToken
}

func getToken() (string, error) {
	config := config{
		clientID:     "nkvyVWVBiqOKs_o7dLkUF2KHv2Ya",
		clientSecret: "HR_ssS_Kv1q_9xq1j_wJr1F8Fn0a", // i'm allowed to do this :)
		redirect:     "http://localhost:7444",
		port:         7444,
	}

	var toks authCodeResponse
	var err error

	if len(viper.GetString("wso2.refresh-token")) > 0 {
		// get a new token
		toks, err = getTokens("refresh", viper.GetString("wso2.refresh-token"), config)
		if err != nil {
			if strings.Contains(err.Error(), "Provided Authorization Grant is invalid.") {
				// invalidate the current refresh token, it's probably invalid
				viper.Set("wso2.refresh-token", "")
				viper.WriteConfig()
			} else {
				return "", fmt.Errorf("unable to get tokens: %s", err)
			}
		}
	}

	if len(toks.AccessToken) == 0 {
		// get an auth code
		code := getAuthCode(config)

		// get the refresh token
		toks, err = getTokens("authcode", code, config)
		if err != nil {
			return "", fmt.Errorf("unable to get tokens: %s", err)
		}
	}

	if len(toks.RefreshToken) > 0 {
		viper.Set("wso2.refresh-token", toks.RefreshToken)
	}

	err = viper.WriteConfig()
	if err != nil {
		fmt.Printf("unable to save refresh token: %s\n", err)
	}

	return toks.AccessToken, nil
}

func getAuthCode(config config) string {
	codeChan := make(chan string)

	url := fmt.Sprintf("https://api.byu.edu/authorize?response_type=code&client_id=%s&redirect_uri=%s&scope=openid", config.clientID, config.redirect)

	// run the server
	go func() {
		stop := make(chan struct{})
		srv := http.Server{
			Addr: fmt.Sprintf(":%v", config.port),
		}

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			code, ok := r.URL.Query()["code"]
			if !ok {
				io.WriteString(w, fmt.Sprintf(`
				<html>
					<script>
						window.onload = function() {
							window.location.replace("%s")
						}
					</script>
					<body>
						<span>no auth code found. please retry</span>
					</body>
				</html>
				`, url))
				return
			}

			io.WriteString(w, `
			<html>
				<script>
					window.onload = function() {
						window.close();
					}
				</script>
				<body>
					<span>success. you can close this window</span>
				</body>
			</html>
			`)
			codeChan <- code[len(code)-1]
			stop <- struct{}{}
		})

		go srv.ListenAndServe()
		<-stop
		srv.Close()
	}()

	openBrowser(url)

	code := <-codeChan
	return code
}

func getTokens(method, auth string, config config) (authCodeResponse, error) {
	ret := authCodeResponse{}
	data := url.Values{}

	switch method {
	case "authcode":
		data.Set("grant_type", "authorization_code")
		data.Set("code", auth)
		data.Set("redirect_uri", config.redirect)
	case "refresh":
		data.Set("grant_type", "refresh_token")
		data.Set("refresh_token", auth)
	default:
		return ret, errors.New("Invalid Method")
	}

	req, err := http.NewRequest("POST", "https://api.byu.edu/token", strings.NewReader(data.Encode()))
	if err != nil {
		return ret, fmt.Errorf("unable to build request: %s", err)
	}

	req.SetBasicAuth(config.clientID, config.clientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ret, fmt.Errorf("unable to make request: %s", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ret, fmt.Errorf("unable to read response: %s", err)
	}

	if resp.StatusCode != 200 {
		return ret, fmt.Errorf("non-200 response: %s", body)
	}

	err = json.Unmarshal(body, &ret)
	if err != nil {
		return ret, fmt.Errorf("unable to unmarshal response: %s", err)
	}

	return ret, nil
}

func openBrowser(url string) {
	fmt.Printf("opening %s in the background\n", color.BlueString(url))

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
}
