package avcli

import "context"

type DataService interface {
	Room(context.Context, string) ([]Pi, error)
	Device(context.Context, string) (Pi, error)
}

type Pi struct {
	ID      string
	Address string
	Type    string
}
