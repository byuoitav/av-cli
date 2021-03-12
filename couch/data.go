package couch

import (
	"context"
	"errors"
	"fmt"

	avcli "github.com/byuoitav/av-cli"
)

type device struct {
	ID      string `json:"_id"`
	Address string `json:"address"`
	Type    struct {
		ID string `json:"_id"`
	} `json:"type"`
}

func (d *DataService) prefixQuery(ctx context.Context, prefix string) ([]avcli.Pi, error) {
	db := d.client.DB(ctx, d.database)
	query := map[string]interface{}{
		"selector": map[string]interface{}{
			"_id": map[string]interface{}{
				"$regex": prefix,
			},
			"type._id": map[string]interface{}{
				"$regex": "(Pi3)", // TODO add other types
			},
		},
	}

	rows, err := db.Find(ctx, query)
	if err != nil {
		return []avcli.Pi{}, fmt.Errorf("unable to find: %w", err)
	}

	var pis []avcli.Pi
	for rows.Next() {
		var dev device
		if err := rows.ScanDoc(&dev); err != nil {
			continue
		}

		if !dev.isPi() {
			continue
		}

		pis = append(pis, dev.convert())
	}

	return pis, nil
}

func (d *DataService) Building(ctx context.Context, id string) ([]avcli.Pi, error) {
	return d.prefixQuery(ctx, id+"-")
}

func (d *DataService) Room(ctx context.Context, id string) ([]avcli.Pi, error) {
	return d.prefixQuery(ctx, id+"-")
}

func (d *DataService) Device(ctx context.Context, id string) (avcli.Pi, error) {
	var dev device

	db := d.client.DB(ctx, d.database)
	if err := db.Get(ctx, id).ScanDoc(&dev); err != nil {
		return avcli.Pi{}, fmt.Errorf("unable to get/scan device: %w", err)
	}

	if !dev.isPi() {
		return avcli.Pi{}, errors.New("not a valid pi id")
	}

	return dev.convert(), nil
}

func (d device) convert() avcli.Pi {
	return avcli.Pi{
		ID:      d.ID,
		Address: d.Address,
		Type:    d.Type.ID,
	}
}

// TODO add other types
func (d device) isPi() bool {
	switch d.Type.ID {
	case "Pi3":
		return true
	}

	return false
}
