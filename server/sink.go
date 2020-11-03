package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	avcli "github.com/byuoitav/av-cli"
	"golang.org/x/crypto/ssh"
)

func (s *Server) Sink(id *avcli.ID, stream avcli.AvCli_SinkServer) error {
	ctx, cancel := context.WithTimeout(stream.Context(), 15*time.Second)
	defer cancel()

	return s.runPerPi(ctx, id, stream, func(pi avcli.Pi) error {
		client, err := s.piSSH(ctx, pi.Address)
		if err != nil {
			return fmt.Errorf("unable to ssh: %w", err)
		}
		defer client.Close()

		session, err := client.NewSession()
		if err != nil {
			return fmt.Errorf("unable to open new ssh session: %w", err)
		}
		defer session.Close()

		buf := &bytes.Buffer{}
		session.Stderr = buf
		session.Stdout = buf

		if err := session.Start("sudo reboot"); err != nil {
			return fmt.Errorf("unable to start command: %w", err)
		}

		errResp := make(chan error)
		go func() {
			errResp <- session.Wait()
		}()

		select {
		case err := <-errResp:
			switch {
			case errors.Is(err, &ssh.ExitMissingError{}):
				return nil
			case err != nil:
				return fmt.Errorf("unable to run command: %w. output: %s", err, buf.String())
			}

			return fmt.Errorf("unexpected response from command: %s", buf.String())
		case <-ctx.Done():
			return fmt.Errorf("unable to run command: %w", ctx.Err())
		}
	})
}
