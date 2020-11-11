package couch

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/go-kivik/kivik"
)

type statusErr interface {
	StatusCode() int
}

func (d *DataService) CopyRoom(ctx context.Context, src, dst string) error {
	rooms := d.client.DB(ctx, "rooms")
	uiConfigs := d.client.DB(ctx, "ui-configuration")
	devices := d.client.DB(ctx, d.database)

	// make sure all of the dst docs don't exist
	exists, err := docExists(ctx, rooms, dst)
	switch {
	case err != nil:
		return err
	case exists:
		return fmt.Errorf("%q already exists", dst)
	}

	exists, err = docExists(ctx, uiConfigs, dst)
	switch {
	case err != nil:
		return err
	case exists:
		return fmt.Errorf("%q already exists", dst)
	}

	deviceIDs := make(map[string]string)
	query := map[string]interface{}{
		"selector": map[string]interface{}{
			"_id": map[string]interface{}{
				"$regex": src,
			},
		},
	}

	rows, err := devices.Find(ctx, query)
	if err != nil {
		return err
	}

	for rows.Next() {
		srcDoc := struct {
			ID string `json:"_id"`
		}{}
		if err := rows.ScanDoc(&srcDoc); err != nil {
			return err
		}

		deviceIDs[srcDoc.ID] = strings.ReplaceAll(srcDoc.ID, src, dst)

		exists, err = docExists(ctx, devices, deviceIDs[srcDoc.ID])
		switch {
		case err != nil:
			return err
		case exists:
			return fmt.Errorf("%q already exists", deviceIDs[srcDoc.ID])
		}
	}

	replacements := []replacement{
		{
			old: []byte(src),
			new: []byte(dst),
		},
		{
			// TODO should this logic be somewhere else?
			old: []byte(strings.Replace(src, "-", " ", 1)),
			new: []byte(strings.Replace(dst, "-", " ", 1)),
		},
	}

	if err := copyDoc(ctx, rooms, src, dst, replacements...); err != nil {
		return err
	}

	if err := copyDoc(ctx, uiConfigs, src, dst, replacements...); err != nil {
		return err
	}

	for srcID, dstID := range deviceIDs {
		if err := copyDoc(ctx, devices, srcID, dstID, replacements...); err != nil {
			return err
		}
	}

	return nil
}

type replacement struct {
	old []byte
	new []byte
}

func docExists(ctx context.Context, db *kivik.DB, id string) (bool, error) {
	row := db.Get(ctx, id)
	if statusErr, ok := row.Err.(statusErr); ok && statusErr.StatusCode() == http.StatusNotFound {
		return false, nil
	} else if row.Err != nil {
		return false, row.Err
	}

	return true, nil
}

func copyDoc(ctx context.Context, db *kivik.DB, srcID, dstID string, replacements ...replacement) error {
	row := db.Get(ctx, dstID)
	if statusErr, ok := row.Err.(statusErr); ok && statusErr.StatusCode() == http.StatusNotFound {
	} else if row.Err != nil {
		return row.Err
	}

	/* this updates a previously existing document
	dst := dstID
	if row.Rev != "" {
		dst = dstID + "?rev=" + row.Rev
	}
	*/

	if _, err := db.Copy(ctx, dstID, srcID); err != nil {
		return err
	}

	// make the replacements to the doc
	row = db.Get(ctx, dstID)
	if row.Err != nil {
		return row.Err
	}

	doc, err := ioutil.ReadAll(db.Get(ctx, dstID).Body)
	if err != nil {
		return err
	}

	for _, replace := range replacements {
		doc = bytes.ReplaceAll(doc, replace.old, replace.new)
	}

	// upload the replaced doc
	if _, err := db.Put(ctx, dstID, doc); err != nil {
		return err
	}

	return nil
}

// TODO change to Copy()
