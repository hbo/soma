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
	"github.com/mjolnir42/soma/internal/super"
	"github.com/mjolnir42/soma/lib/proto"
)

// AttributeList function
func AttributeList(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer PanicCatcher(w)

	if !super.IsAuthorized(&msg.Authorization{
		AuthUser:   params.ByName(`AuthenticatedUser`),
		RemoteAddr: extractAddress(r.RemoteAddr),
		Section:    `attribute`,
		Action:     `list`,
	}) {
		DispatchForbidden(&w, nil)
		return
	}

	returnChannel := make(chan msg.Result)
	handler := handlerMap[`attribute_r`].(*attributeRead)
	handler.input <- msg.Request{
		Section:    `attribute`,
		Action:     `list`,
		Reply:      returnChannel,
		RemoteAddr: extractAddress(r.RemoteAddr),
		AuthUser:   params.ByName(`AuthenticatedUser`),
	}
	result := <-returnChannel
	SendMsgResult(&w, &result)
}

// AttributeShow function
func AttributeShow(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer PanicCatcher(w)

	if !super.IsAuthorized(&msg.Authorization{
		AuthUser:   params.ByName(`AuthenticatedUser`),
		RemoteAddr: extractAddress(r.RemoteAddr),
		Section:    `attribute`,
		Action:     `show`,
	}) {
		DispatchForbidden(&w, nil)
		return
	}

	returnChannel := make(chan msg.Result)
	handler := handlerMap[`attribute_r`].(*attributeRead)
	handler.input <- msg.Request{
		Section:    `attribute`,
		Action:     `show`,
		Reply:      returnChannel,
		RemoteAddr: extractAddress(r.RemoteAddr),
		AuthUser:   params.ByName(`AuthenticatedUser`),
		Attribute: proto.Attribute{
			Name: params.ByName(`attribute`),
		},
	}
	result := <-returnChannel
	SendMsgResult(&w, &result)
}

// AttributeAdd function
func AttributeAdd(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer PanicCatcher(w)

	if !super.IsAuthorized(&msg.Authorization{
		AuthUser:   params.ByName(`AuthenticatedUser`),
		RemoteAddr: extractAddress(r.RemoteAddr),
		Section:    `attribute`,
		Action:     `add`,
	}) {
		DispatchForbidden(&w, nil)
		return
	}

	cReq := proto.NewStateRequest()
	if err := DecodeJsonBody(r, &cReq); err != nil {
		DispatchBadRequest(&w, err)
		return
	}

	returnChannel := make(chan msg.Result)
	handler := handlerMap[`attribute_w`].(*attributeWrite)
	handler.input <- msg.Request{
		Section:    `attribute`,
		Action:     `add`,
		Reply:      returnChannel,
		RemoteAddr: extractAddress(r.RemoteAddr),
		AuthUser:   params.ByName(`AuthenticatedUser`),
		Attribute: proto.Attribute{
			Name:        cReq.Attribute.Name,
			Cardinality: cReq.Attribute.Cardinality,
		},
	}
	result := <-returnChannel
	SendMsgResult(&w, &result)
}

// AttributeRemove function
func AttributeRemove(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer PanicCatcher(w)

	if !super.IsAuthorized(&msg.Authorization{
		AuthUser:   params.ByName(`AuthenticatedUser`),
		RemoteAddr: extractAddress(r.RemoteAddr),
		Section:    `attribute`,
		Action:     `remove`,
	}) {
		DispatchForbidden(&w, nil)
		return
	}

	returnChannel := make(chan msg.Result)
	handler := handlerMap[`attribute_w`].(*attributeWrite)
	handler.input <- msg.Request{
		Section:    `attribute`,
		Action:     `remove`,
		Reply:      returnChannel,
		RemoteAddr: extractAddress(r.RemoteAddr),
		AuthUser:   params.ByName(`AuthenticatedUser`),
		Attribute: proto.Attribute{
			Name: params.ByName(`attribute`),
		},
	}
	result := <-returnChannel
	SendMsgResult(&w, &result)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
