package adm

import (
	"encoding/json"
	"fmt"

	"github.com/mjolnir42/soma/lib/auth"
	"gopkg.in/resty.v1"
)

func KeyExchange(c *resty.Client) (*auth.Kex, error) {
	var (
		err       error
		kex, peer *auth.Kex
		resp      *resty.Response
	)

	kex = auth.NewKex()
	kex.SetTimeUTC()
	if resp, err = c.R().SetBody(kex).Post(`/kex/`); err != nil {
		goto fail
	} else if resp.StatusCode() != 200 {
		err = fmt.Errorf("Incorrect response code from SOMA, expected 200, got: %d",
			resp.StatusCode())
		goto fail
	}

	peer = &auth.Kex{}
	if err = json.Unmarshal(resp.Body(), peer); err != nil {
		goto fail
	}

	// store settings from peer
	kex.SetPeerKey(peer.PublicKey())
	kex.ImportInitializationVector(peer.ExportInitializationVector())
	if err = kex.SetRequestUUID(peer.Request.String()); err != nil {
		goto fail
	}

	return kex, nil

fail:
	return nil, err
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
