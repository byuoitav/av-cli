package cas

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Login .
func Login(ctx context.Context, username, password string) error {
	loginreq := loginRequest{
		Username: username,
		Password: password,
	}

	body, err := json.Marshal(loginreq)
	if err != nil {
		return fmt.Errorf("unable to marshal login request: %s", err)
	}

	req, err := http.NewRequest("POST", "http://cas.byu.edu/cas/login", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("unable to build login http request: %s", err)
	}

	req = req.WithContext(ctx)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to send login request: %s", err)
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("unable to read response: %s", err)
	}


	return nil
}
