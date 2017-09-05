package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"unicode/utf8"

	"github.com/julienschmidt/httprouter"
	"github.com/mjolnir42/soma/internal/msg"
	"github.com/mjolnir42/soma/internal/super"
	"github.com/mjolnir42/soma/lib/proto"
)

// RepositoryList function
func RepositoryList(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer PanicCatcher(w)

	if !super.IsAuthorized(&msg.Authorization{
		AuthUser:   params.ByName(`AuthenticatedUser`),
		RemoteAddr: extractAddress(r.RemoteAddr),
		Section:    `repository`,
		Action:     `list`,
	}) {
		DispatchForbidden(&w, nil)
		return
	}

	returnChannel := make(chan somaResult)
	handler := handlerMap["repositoryReadHandler"].(*somaRepositoryReadHandler)
	handler.input <- somaRepositoryRequest{
		action: "list",
		reply:  returnChannel,
	}
	result := <-returnChannel

	// declare here since goto does not jump over declarations
	cReq := proto.NewRepositoryFilter()
	if result.Failure() {
		goto skip
	}

	_ = DecodeJsonBody(r, &cReq)
	if cReq.Filter.Repository.Name != "" {
		filtered := []somaRepositoryResult{}
		for _, i := range result.Repositories {
			if i.Repository.Name == cReq.Filter.Repository.Name {
				filtered = append(filtered, i)
			}
		}
		result.Repositories = filtered
	}

skip:
	SendRepositoryReply(&w, &result)
}

// RepositoryShow function
func RepositoryShow(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer PanicCatcher(w)

	if !super.IsAuthorized(&msg.Authorization{
		AuthUser:   params.ByName(`AuthenticatedUser`),
		RemoteAddr: extractAddress(r.RemoteAddr),
		Section:    `repository`,
		Action:     `show`,
	}) {
		DispatchForbidden(&w, nil)
		return
	}

	returnChannel := make(chan somaResult)
	handler := handlerMap["repositoryReadHandler"].(*somaRepositoryReadHandler)
	handler.input <- somaRepositoryRequest{
		action: "show",
		reply:  returnChannel,
		Repository: proto.Repository{
			Id: params.ByName("repository"),
		},
	}
	result := <-returnChannel
	SendRepositoryReply(&w, &result)
}

// RepositoryCreate function
func RepositoryCreate(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer PanicCatcher(w)

	if !super.IsAuthorized(&msg.Authorization{
		AuthUser:   params.ByName(`AuthenticatedUser`),
		RemoteAddr: extractAddress(r.RemoteAddr),
		Section:    `repository`,
		Action:     `create`,
	}) {
		DispatchForbidden(&w, nil)
		return
	}

	cReq := proto.NewRepositoryRequest()
	err := DecodeJsonBody(r, &cReq)
	if err != nil {
		DispatchBadRequest(&w, err)
		return
	}

	nameLen := utf8.RuneCountInString(cReq.Repository.Name)
	if nameLen < 4 || nameLen > 128 {
		DispatchBadRequest(&w, fmt.Errorf(`Illegal repository name length (4 < x <= 128)`))
		return
	}

	returnChannel := make(chan somaResult)
	handler := handlerMap["forestCustodian"].(*forestCustodian)
	handler.input <- somaRepositoryRequest{
		action:     "add",
		reply:      returnChannel,
		remoteAddr: extractAddress(r.RemoteAddr),
		user:       params.ByName(`AuthenticatedUser`),
		Repository: proto.Repository{
			Name:      cReq.Repository.Name,
			TeamId:    cReq.Repository.TeamId,
			IsDeleted: cReq.Repository.IsDeleted,
			IsActive:  cReq.Repository.IsActive,
		},
	}
	result := <-returnChannel
	SendRepositoryReply(&w, &result)
}

// RepositoryAddProperty function
func RepositoryAddProperty(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer PanicCatcher(w)

	if !super.IsAuthorized(&msg.Authorization{
		AuthUser:     params.ByName(`AuthenticatedUser`),
		RemoteAddr:   extractAddress(r.RemoteAddr),
		Section:      `repository`,
		Action:       `add_property`,
		RepositoryID: params.ByName(`repository`),
	}) {
		DispatchForbidden(&w, nil)
		return
	}

	cReq := proto.NewRepositoryRequest()
	if err := DecodeJsonBody(r, &cReq); err != nil {
		DispatchBadRequest(&w, err)
		return
	}
	switch {
	case params.ByName("repository") != cReq.Repository.Id:
		DispatchBadRequest(&w,
			fmt.Errorf("Mismatched repository ids: %s, %s",
				params.ByName("repository"),
				cReq.Repository.Id))
		return
	case len(*cReq.Repository.Properties) != 1:
		DispatchBadRequest(&w,
			fmt.Errorf("Expected property count 1, actual count: %d",
				len(*cReq.Repository.Properties)))
		return
	case params.ByName("type") != (*cReq.Repository.Properties)[0].Type:
		DispatchBadRequest(&w,
			fmt.Errorf("Mismatched property types: %s, %s",
				params.ByName("type"),
				(*cReq.Repository.Properties)[0].Type))
		return
	case (params.ByName("type") == "service") && (*cReq.Repository.Properties)[0].Service.Name == "":
		DispatchBadRequest(&w,
			fmt.Errorf("Empty service name is invalid"))
		return
	}

	returnChannel := make(chan somaResult)
	handler := handlerMap["guidePost"].(*guidePost)
	handler.input <- treeRequest{
		RequestType: "repository",
		Action:      fmt.Sprintf("add_%s_property_to_repository", params.ByName("type")),
		AuthUser:    params.ByName(`AuthenticatedUser`),
		reply:       returnChannel,
		Repository: somaRepositoryRequest{
			action:     fmt.Sprintf("%s_property_new", params.ByName("type")),
			Repository: *cReq.Repository,
		},
	}
	result := <-returnChannel
	SendRepositoryReply(&w, &result)
}

// RepositoryRemoveProperty function
func RepositoryRemoveProperty(w http.ResponseWriter, r *http.Request,
	params httprouter.Params) {
	defer PanicCatcher(w)

	if !super.IsAuthorized(&msg.Authorization{
		AuthUser:     params.ByName(`AuthenticatedUser`),
		RemoteAddr:   extractAddress(r.RemoteAddr),
		Section:      `repository`,
		Action:       `remove_property`,
		RepositoryID: params.ByName(`repository`),
	}) {
		DispatchForbidden(&w, nil)
		return
	}

	repo := &proto.Repository{
		Id: params.ByName(`repository`),
		Properties: &[]proto.Property{
			proto.Property{
				Type:             params.ByName(`type`),
				RepositoryId:     params.ByName(`repository`),
				SourceInstanceId: params.ByName(`source`),
			},
		},
	}

	returnChannel := make(chan somaResult)
	handler := handlerMap["guidePost"].(*guidePost)
	handler.input <- treeRequest{
		RequestType: "repository",
		Action: fmt.Sprintf("delete_%s_property_from_repository",
			params.ByName("type")),
		AuthUser: params.ByName(`AuthenticatedUser`),
		reply:    returnChannel,
		Repository: somaRepositoryRequest{
			action: fmt.Sprintf("%s_property_remove",
				params.ByName("type")),
			Repository: *repo,
		},
	}
	result := <-returnChannel
	SendRepositoryReply(&w, &result)
}

// SendRepositoryReply function
func SendRepositoryReply(w *http.ResponseWriter, r *somaResult) {
	result := proto.NewRepositoryResult()
	if r.MarkErrors(&result) {
		goto dispatch
	}
	for _, i := range (*r).Repositories {
		*result.Repositories = append(*result.Repositories, i.Repository)
		if i.ResultError != nil {
			*result.Errors = append(*result.Errors, i.ResultError.Error())
		}
	}

dispatch:
	json, err := json.Marshal(result)
	if err != nil {
		DispatchInternalError(w, err)
		return
	}
	DispatchJsonReply(w, &json)
	return
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
