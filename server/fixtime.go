package server

import (
	"bytes"
	"context"
	"fmt"
	"time"

	avcli "github.com/byuoitav/av-cli"
)

func (s *Server) FixTime(id *avcli.ID, stream avcli.AvCli_FixTimeServer) error {
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

		if err := session.Start("date && sudo ntpdate tick.byu.edu && date"); err != nil {
			return fmt.Errorf("unable to start command: %w", err)
		}

		errResp := make(chan error)
		go func() {
			errResp <- session.Wait()
		}()

		select {
		case err := <-errResp:
			if err != nil {
				return fmt.Errorf("unable to run command: %w. output: %s", err, buf.String())
			}

			return nil
		case <-ctx.Done():
			return fmt.Errorf("unable to run command: %w", ctx.Err())
		}
	})
}
