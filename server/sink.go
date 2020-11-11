package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	avcli "github.com/byuoitav/av-cli"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
)

func (s *Server) Sink(id *avcli.ID, stream avcli.AvCli_SinkServer) error {
	log := s.log(stream.Context())
	log.Info("Sinking", zap.String("id", id.GetId()))

	ctx, cancel := context.WithTimeout(stream.Context(), 15*time.Second)
	defer cancel()

	return s.runPerPi(ctx, id, stream, func(pi avcli.Pi) error {
		log := log.With(zap.String("pi", pi.ID))

		client, err := s.piSSH(ctx, pi.Address)
		if err != nil {
			log.Warn("unable to ssh", zap.Error(err))
			return fmt.Errorf("unable to ssh: %w", err)
		}
		defer client.Close()

		session, err := client.NewSession()
		if err != nil {
			log.Warn("unable to open new session", zap.Error(err))
			return fmt.Errorf("unable to open new ssh session: %w", err)
		}
		defer session.Close()

		buf := &bytes.Buffer{}
		session.Stderr = buf
		session.Stdout = buf

		if err := session.Start("sudo reboot"); err != nil {
			log.Warn("unable to start command", zap.Error(err))
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
				log.Info("Successfully sunk")
				return nil
			case err != nil:
				log.Warn("unable to run command", zap.Error(err), zap.String("output", buf.String()))
				return fmt.Errorf("unable to run command: %w. output: %s", err, buf.String())
			}

			log.Warn("unexpected response from command", zap.String("output", buf.String()))
			return fmt.Errorf("unexpected response from command: %s", buf.String())
		case <-ctx.Done():
			log.Warn("unable to run command", zap.Error(ctx.Err()))
			return fmt.Errorf("unable to run command: %w", ctx.Err())
		}
	})
}
