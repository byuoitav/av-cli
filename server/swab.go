package server

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	avcli "github.com/byuoitav/av-cli"
)

func (s *Server) Swab(id *avcli.ID, stream avcli.AvCli_SwabServer) error {
	// TODO handle different types differently
	// TODO status errors
	ctx, cancel := context.WithTimeout(stream.Context(), 15*time.Second)
	defer cancel()

	return s.runPerPi(ctx, id, stream, func(pi avcli.Pi) error {
		// start the replication
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://%s:7012/replication/start", pi.Address), nil)
		if err != nil {
			return fmt.Errorf("unable to build replication request: %w", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("unable to do replication request: %w", err)
		}
		resp.Body.Close()

		// let it replicate for a bit
		select {
		case <-time.After(5 * time.Second):
		case <-ctx.Done():
			return fmt.Errorf("unable to wait for replication: %w", ctx.Err())
		}

		// refresh the ui (TODO check the url)
		req, err = http.NewRequestWithContext(ctx, http.MethodPut, fmt.Sprintf("http://%s:8888/refresh", pi.Address), nil)
		if err != nil {
			return fmt.Errorf("unable to build refresh request: %w", err)
		}

		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			return fmt.Errorf("unable to do refresh request: %w", err)
		}
		resp.Body.Close()

		// restart the dmm
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

		if err := session.Start("sudo systemctl restart device-monitoring.service"); err != nil {
			return fmt.Errorf("unable to restart device monitoring: unable to start command: %w", err)
		}

		errResp := make(chan error)
		go func() {
			errResp <- session.Wait()
		}()

		select {
		case err := <-errResp:
			if err != nil {
				return fmt.Errorf("unable to restart device monitoring: %w. output: %s", err, buf.String())
			}

			return nil
		case <-ctx.Done():
			return fmt.Errorf("unable to restart device monitoring: %w", ctx.Err())
		}
	})
}
