package wso2

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

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

// GetIDToken .
func GetIDToken() string {
	return getToks().IDToken
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
				_ = viper.WriteConfig()
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

	id, err := os.Hostname()
	if err != nil {
		b := make([]byte, 16)
		_, err := rand.Read(b)
		if err != nil {
			log.Fatal(err)
		}
		id = fmt.Sprintf("%x-%x-%x-%x-%x",
			b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
		fmt.Println(id)
	}
	id = url.QueryEscape(id)
	url := fmt.Sprintf("https://api.byu.edu/authorize?response_type=code&client_id=%s&redirect_uri=%s&scope=openid %s", config.clientID, config.redirect, fmt.Sprintf("device_%s", id))

	// run the server
	go func() {
		srv := http.Server{
			Addr: fmt.Sprintf(":%v", config.port),
		}

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			code, ok := r.URL.Query()["code"]
			if !ok {
				_, _ = io.WriteString(w, fmt.Sprintf(`
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

			_, _ = io.WriteString(w, `
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

		go func() {
			_ = srv.ListenAndServe()
		}()

		<-stopSrv
		srv.Close()
	}()

	err = OpenBrowser(url)
	if err != nil {
		stopSrv <- struct{}{}

		go func() {
			fmt.Printf("\n\nunable to open browser: %s\nCopy link below into browser, and paste the auth code from the url.\n%s\n", err, color.New(color.FgBlue, color.Underline, color.Bold).Sprint(url))

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
	switch runtime.GOOS {
	case "linux":
		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		defer cancel()

		var exiterr *exec.ExitError
		cmd := exec.CommandContext(ctx, "xdg-open", url)

		err := cmd.Run()
		if errors.As(err, &exiterr) {
			switch exiterr.ExitCode() {
			case 3:
				return errors.New("no tool found to open url's")
			default:
				return fmt.Errorf("xdg-open failed with status code %d", exiterr.ExitCode())
			}
		}

		return err
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	default:
		return fmt.Errorf("don't know how to open browser on %s", runtime.GOOS)
	}
}
