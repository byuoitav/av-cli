package server

import (
	"context"
	"time"

	"golang.org/x/crypto/ssh"
)

// TODO should change to using an ssh key
func (s *Server) piSSH(ctx context.Context, addr string) (*ssh.Client, error) {
	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(3 * time.Second)
	}

	sshConfig := &ssh.ClientConfig{
		User:            "pi",
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.Password(s.PiPassword),
		},
		Timeout: time.Until(deadline),
	}

	return ssh.Dial("tcp", addr+":22", sshConfig)
}
