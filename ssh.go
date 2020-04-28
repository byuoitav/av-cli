package avcli

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

var (
	defaultKeys = [...]string{"$HOME/.ssh/id_rsa"}
)

// NewSSHClient .
func NewSSHClient(address string) (*ssh.Client, error) {
	label := fmt.Sprintf("%s's password", address)

	sshConfig := &ssh.ClientConfig{
		User:            "pi",
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.PublicKeysCallback(getSigners),
			ssh.RetryableAuthMethod(ssh.PasswordCallback(getPasswordFunc(label, 3)), 3),
		},
		Timeout: 2 * time.Second,
	}

	return ssh.Dial("tcp", address+":22", sshConfig)
}

func getSigners() ([]ssh.Signer, error) {
	var signers []ssh.Signer

	for _, path := range defaultKeys {
		bytes, err := ioutil.ReadFile(path)
		if err != nil {
			continue
		}

		key, err := ssh.ParsePrivateKey(bytes)
		if err != nil {
			continue
		}

		signers = append(signers, key)
	}

	return signers, nil
}
func getPasswordFunc(label string, maxTries int) func() (string, error) {

	return func() (string, error) {
		password := os.Getenv("PI_PASSWORD")
		if len(password) == 0 {
			return "", fmt.Errorf("PI_PASSWORD not set")
		}

		return password, nil
	}
}
