package avcli

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
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
	tries := 0

	return func() (string, error) {
		if tries > 0 {
			fmt.Printf("Permission denied, please try again.\n")
		}

		if tries == 0 && viper.IsSet("pi_password") {
			fmt.Printf("Using configured password.\n")
			return viper.GetString("pi_password"), nil
		}

		tries++

		passPrompt := promptui.Prompt{
			Label: label,
			Mask:  '*',
		}

		pass, err := passPrompt.Run()
		if err != nil {
			return "", err
		}

		return pass, nil
	}
}
