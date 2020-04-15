package main

import (
	"context"
	"net/http"
	"time"

	avcli "github.com/byuoitav/av-cli"
	"github.com/labstack/echo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type slack struct {
	cli   avcli.AvCliClient
	token string
}

type slackRequest struct {
	Token          string `form:"token"`
	Command        string `form:"command"`
	Text           string `form:"text"`
	ResponseURL    string `form:"response_url"`
	TriggerID      string `form:"trigger_id"`
	UserID         string `form:"user_id"`
	UserName       string `form:"user_name"`
	TeamID         string `form:"team_id"`
	TeamName       string `form:"team_name"`
	EnterpriseID   string `form:"enterprise_id"`
	EnterpriseName string `form:"enterprise_name"`
	ChannelID      string `form:"channel_id"`
	ChannelName    string `form:"channel_name"`
}

func (s *slack) handleRequest(c echo.Context) error {
	// TODO validate request came from slack
	// TODO get the user, send in metadata
	// TODO write handler logic here
	// TODO actual error handling for slack API

	var req slackRequest
	if err := c.Bind(&req); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	auth := auth{
		token: s.token,
		netID: req.UserID, // TODO should be their netID
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	result, err := s.cli.Screenshot(ctx, &avcli.ID{Id: req.Text}, grpc.PerRPCCredentials(auth))
	switch {
	case err != nil:
		if s, ok := status.FromError(err); ok {
			switch s.Code() {
			case codes.Unavailable:
				return c.String(http.StatusBadGateway, s.Err().Error())
			case codes.Unauthenticated:
				return c.String(http.StatusForbidden, s.Err().Error())
			default:
				return c.String(http.StatusInternalServerError, s.Err().Error())
			}
		}

		return c.String(http.StatusInternalServerError, err.Error())
	case result == nil:
		return c.String(http.StatusInternalServerError, "this is weird")
	}

	return c.Blob(http.StatusOK, "image/jpeg", result.GetPhoto())
}
