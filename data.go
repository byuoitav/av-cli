package avcli

import "context"

type DataService interface {
	Building(context.Context, string) ([]Pi, error)
	Room(context.Context, string) ([]Pi, error)
	Device(context.Context, string) (Pi, error)
	CopyRoom(ctx context.Context, from, to string) error
}

type Pi struct {
	ID      string
	Address string
	Type    string
}
