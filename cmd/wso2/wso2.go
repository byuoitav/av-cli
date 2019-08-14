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

	"github.com/dgrijalva/jwt-go"
	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
)

var (
	toks tokens
	once sync.Once
)

type config struct {
	clientID     string
	clientSecret string
	redirect     string
	port         int
}

type tokens struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token"`
}

// IDInfo .
type IDInfo struct {
	PersonID           string `json:"person_id"`
	Surname            string `json:"surname"`
	PreferredFirstName string `json:"preferred_first_name"`
	RestOfName         string `json:"rest_of_name"`
	NetID              string `json:"net_id"`
	Suffix             string `json:"suffix"`
	SortName           string `json:"sort_name"`
	Prefix             string `json:"prefix"`
	SurnamePosition    string `json:"surname_position"`
	BYUID              string `json:"byu_id"`

	// because they don't send back a normal jwt..?
	Audience []string `json:"aud,omitempty"`
	jwt.StandardClaims
}

// GetAccessToken .
func GetAccessToken() string {
	return getToks().AccessToken
}

// GetIDInfo .
func GetIDInfo() (*IDInfo, error) {
	id := getToks().IDToken

	parser := &jwt.Parser{
		SkipClaimsValidation: false,
	}

	token, _, err := parser.ParseUnverified(id, &IDInfo{})
	if err != nil {
		return nil, fmt.Errorf("unable to parse jwt: %s", err)
	}

	if claims, ok := token.Claims.(*IDInfo); ok {
		return claims, nil
	}

	return nil, fmt.Errorf("claims not found in jwt. claims: %+v", token.Claims)
}

func getToks() tokens {
	once.Do(func() {
		var err error
		toks, err = getTokens()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	})

	return toks
}

func getTokens() (tokens, error) {
	config := config{
		clientID:     "nkvyVWVBiqOKs_o7dLkUF2KHv2Ya",
		clientSecret: "HR_ssS_Kv1q_9xq1j_wJr1F8Fn0a", // i'm allowed to do this ;)
		redirect:     "http://localhost:7444",
		port:         7444,
	}

	var err error

	if len(viper.GetString("wso2.refresh-token")) > 0 {
		// get a new token
		toks, err = doTokenRequest("refresh", viper.GetString("wso2.refresh-token"), config)
		if err != nil {
			if strings.Contains(err.Error(), "Provided Authorization Grant is invalid.") {
				// invalidate the current refresh token, it's probably invalid
				viper.Set("wso2.refresh-token", "")
				viper.WriteConfig()
			} else {
				return tokens{}, fmt.Errorf("unable to get tokens: %s", err)
			}
		}
	}

	if len(toks.AccessToken) == 0 {
		// get an auth code
		code := getAuthCode(config)

		// get the refresh token
		toks, err = doTokenRequest("authcode", code, config)
		if err != nil {
			return tokens{}, fmt.Errorf("unable to get tokens: %s", err)
		}
	}

	if len(toks.RefreshToken) > 0 {
		viper.Set("wso2.refresh-token", toks.RefreshToken)
	}

	err = viper.WriteConfig()
	if err != nil {
		fmt.Printf("unable to save refresh token: %s\n", err)
	}

	return toks, nil
}

func getAuthCode(config config) string {
	codeChan := make(chan string)
	stopSrv := make(chan struct{})

	url := fmt.Sprintf("https://api.byu.edu/authorize?response_type=code&client_id=%s&redirect_uri=%s&scope=openid", config.clientID, config.redirect)

	// run the server
	go func() {
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
			stopSrv <- struct{}{}
		})

		go srv.ListenAndServe()
		<-stopSrv
		srv.Close()
	}()

	err := OpenBrowser(url)
	if err != nil {
		stopSrv <- struct{}{}

		go func() {
			fmt.Printf("Unable to open browser: %s. Copy link below into browser, and paste the auth code from the url.\n%s\n", err, color.New(color.FgBlue, color.Underline, color.Bold).Sprint(url))

			codePrompt := promptui.Prompt{
				Label: "Auth Code from URL",
			}

			c, err := codePrompt.Run()
			if err != nil {
				fmt.Printf("unable to get auth code: %s\n", err)
				os.Exit(1)
			}

			codeChan <- c
		}()
	}

	code := <-codeChan
	return code
}

func doTokenRequest(method, auth string, config config) (tokens, error) {
	ret := tokens{}
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

// OpenBrowser .
func OpenBrowser(url string) error {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("don't know how to open browser on %s", runtime.GOOS)
	}

	if err != nil {
		return err
	}

	return nil
}
