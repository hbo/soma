/*-
 * Copyright (c) 2016, 1&1 Internet SE
 * Copyright (c) 2016, Jörg Pernfuß
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/mjolnir42/soma/internal/msg"
	"github.com/mjolnir42/soma/lib/proto"
)

// InstanceShow returns information about a check instance
func InstanceShow(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer PanicCatcher(w)

	if !fixIsAuthorized(&msg.Authorization{
		AuthUser:   params.ByName(`AuthenticatedUser`),
		RemoteAddr: extractAddress(r.RemoteAddr),
		Section:    `instance`,
		Action:     `show`,
	}) {
		DispatchForbidden(&w, nil)
		return
	}

	returnChannel := make(chan msg.Result)
	handler := handlerMap[`instance_r`].(*instance)
	handler.input <- msg.Request{
		Section:    `instance`,
		Action:     `show`,
		Reply:      returnChannel,
		RemoteAddr: extractAddress(r.RemoteAddr),
		AuthUser:   params.ByName(`AuthenticatedUser`),
		Instance: proto.Instance{
			ID: params.ByName(`instance`),
		},
	}
	result := <-returnChannel
	SendMsgResult(&w, &result)
}

// InstanceVersions returns information about a check instance's
// version history
func InstanceVersions(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer PanicCatcher(w)

	if !fixIsAuthorized(&msg.Authorization{
		AuthUser:   params.ByName(`AuthenticatedUser`),
		RemoteAddr: extractAddress(r.RemoteAddr),
		Section:    `instance`,
		Action:     `versions`,
	}) {
		DispatchForbidden(&w, nil)
		return
	}

	returnChannel := make(chan msg.Result)
	handler := handlerMap[`instance_r`].(*instance)
	handler.input <- msg.Request{
		Section:    `instance`,
		Action:     `versions`,
		Reply:      returnChannel,
		RemoteAddr: extractAddress(r.RemoteAddr),
		AuthUser:   params.ByName(`AuthenticatedUser`),
		Instance: proto.Instance{
			ID: params.ByName(`instance`),
		},
	}
	result := <-returnChannel
	SendMsgResult(&w, &result)
}

// InstanceList returns the list of instances in the subtree
// below the queried object.
// Currently only supports repositories and buckets as target.
func InstanceList(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer PanicCatcher(w)

	if !fixIsAuthorized(&msg.Authorization{
		AuthUser:     params.ByName(`AuthenticatedUser`),
		RemoteAddr:   extractAddress(r.RemoteAddr),
		Section:      `instance`,
		Action:       `list`,
		RepositoryID: params.ByName(`repository`),
		BucketID:     params.ByName(`bucket`),
	}) {
		DispatchForbidden(&w, nil)
		return
	}

	listT := ``
	switch {
	case params.ByName(`repository`) != ``:
		listT = `repository`
	case params.ByName(`bucket`) != ``:
		listT = `bucket`
	case params.ByName(`group`) != ``:
		fallthrough
	case params.ByName(`cluster`) != ``:
		fallthrough
	case params.ByName(`node`) != ``:
		DispatchNotImplemented(&w, nil)
		return
	default:
		DispatchBadRequest(&w, nil)
		return
	}

	returnChannel := make(chan msg.Result)
	handler := handlerMap[`instance_r`].(*instance)
	handler.input <- msg.Request{
		Section:    `instance`,
		Action:     `list`,
		Reply:      returnChannel,
		RemoteAddr: extractAddress(r.RemoteAddr),
		AuthUser:   params.ByName(`AuthenticatedUser`),
		Instance: proto.Instance{
			ObjectID:   params.ByName(listT),
			ObjectType: listT,
		},
	}
	result := <-returnChannel
	SendMsgResult(&w, &result)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
