package server

import (
	"bytes"
	"context"
	"fmt"
	"time"

	avcli "github.com/byuoitav/av-cli"
	"google.golang.org/grpc/status"
)

func (s *Server) FixTime(id *avcli.ID, stream avcli.AvCli_FixTimeServer) error {
	ctx, cancel := context.WithTimeout(stream.Context(), 15*time.Second)
	defer cancel()

	pis, err := s.Pis(ctx, id)
	if err != nil {
		return err
	}

	results := make(chan *avcli.IDResult)
	defer close(results)

	for i := range pis {
		go func(pi avcli.Pi) {
			var errstr string
			if err := s.fixTime(ctx, pi); err != nil {
				errstr = err.Error()
			}

			results <- &avcli.IDResult{
				Id:    pi.ID,
				Error: errstr,
			}
		}(pis[i])
	}

	expectedResults := len(pis)

	for {
		select {
		case result := <-results:
			if err := stream.Send(result); err != nil {
				return fmt.Errorf("unable to send result: %w", err)
			}

			expectedResults--
			if expectedResults == 0 {
				return nil
			}
		case <-ctx.Done():
			return status.FromContextError(ctx.Err()).Err()
		}
	}
}

func (s *Server) fixTime(ctx context.Context, pi avcli.Pi) error {
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
}
