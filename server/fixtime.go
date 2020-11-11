package server

import (
	"bytes"
	"context"
	"fmt"
	"time"

	avcli "github.com/byuoitav/av-cli"
	"go.uber.org/zap"
)

func (s *Server) FixTime(id *avcli.ID, stream avcli.AvCli_FixTimeServer) error {
	log := s.log(stream.Context())
	log.Info("Fixing time", zap.String("id", id.GetId()))

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

		if err := session.Start("date && sudo ntpdate tick.byu.edu && date"); err != nil {
			log.Warn("unable to start command", zap.Error(err))
			return fmt.Errorf("unable to start command: %w", err)
		}

		errResp := make(chan error)
		go func() {
			errResp <- session.Wait()
		}()

		select {
		case err := <-errResp:
			if err != nil {
				log.Warn("unable to run command", zap.Error(err), zap.String("output", buf.String()))
				return fmt.Errorf("unable to run command: %w. output: %s", err, buf.String())
			}

			log.Info("Successfully fixed time")
			return nil
		case <-ctx.Done():
			log.Warn("unable to run command", zap.Error(ctx.Err()))
			return fmt.Errorf("unable to run command: %w", ctx.Err())
		}
	})
}
