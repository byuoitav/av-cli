package avcli

import (
	context "context"
	fmt "fmt"
	"io/ioutil"
	"net/http"

	empty "github.com/golang/protobuf/ptypes/empty"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

//go:generate protoc -I ./ --go_out=plugins=grpc:./ ./av-cli.proto

type Server struct {
	Logger Logger
}

func (s *Server) Swab(*ID, AvCli_SwabServer) error {
	return status.Errorf(codes.Unimplemented, "method Swab not implemented")
}

func (s *Server) Float(*ID, AvCli_FloatServer) error {
	return status.Errorf(codes.Unimplemented, "method Float not implemented")
}

func (s *Server) Screenshot(ctx context.Context, id *ID) (*ScreenshotResult, error) {
	// TODO validate id
	// TODO lookup id in database, use that address

	handle := func(err error) error {
		s.warnf(err.Error())
		return err
	}

	s.infof("Taking a screenshot of %q", id.GetId())
	url := fmt.Sprintf("http://%s.byu.edu:10000/device/screenshot", id.GetId())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, handle(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, handle(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, handle(err)
	}

	switch resp.StatusCode {
	case http.StatusInternalServerError:
		return nil, handle(fmt.Errorf("failed to get screenshot: %s", body))
	}

	return &ScreenshotResult{
		Photo: body,
	}, nil
}

func (s *Server) DuplicateRoom(context.Context, *DuplicateRoomRequest) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DuplicateRoom not implemented")
}
