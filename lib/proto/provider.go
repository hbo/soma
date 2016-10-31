/*-
 * Copyright (c) 2016, 1&1 Internet SE
 * Copyright (c) 2016, Jörg Pernfuß <joerg.pernfuss@1und1.de>
 * All rights reserved
 *
 * Use of this source code is governed by a 2-clause BSD license
 * that can be found in the LICENSE file.
 */

package proto

type Provider struct {
	Name    string           `json:"name,omitempty"`
	Details *ProviderDetails `json:"details,omitempty"`
}

type ProviderFilter struct {
	Name string `json:"name,omitempty"`
}

type ProviderDetails struct {
	DetailsCreation
}

func NewProviderRequest() Request {
	return Request{
		Flags:    &Flags{},
		Provider: &Provider{},
	}
}

func NewProviderFilter() Request {
	return Request{
		Filter: &Filter{
			Provider: &ProviderFilter{},
		},
	}
}

func NewProviderResult() Result {
	return Result{
		Errors:    &[]string{},
		Providers: &[]Provider{},
	}
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix
