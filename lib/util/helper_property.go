package util

import (
	"fmt"

	"github.com/1and1/soma/lib/proto"
	"gopkg.in/resty.v0"
)

func (u SomaUtil) TryGetCustomPropertyByUUIDOrName(c *resty.Client, s string, r string) string {
	if u.IsUUID(s) {
		return s
	}
	return u.GetPropertyIdByName(c, "custom", s, r)
}

func (u SomaUtil) TryGetServicePropertyByUUIDOrName(c *resty.Client, s string, t string) string {
	if u.IsUUID(s) {
		return s
	}
	return u.GetPropertyIdByName(c, "service", s, t)
}

func (u SomaUtil) TryGetSystemPropertyByUUIDOrName(c *resty.Client, s string) string {
	if u.IsUUID(s) {
		return s
	}
	return u.GetPropertyIdByName(c, "system", s, "none")
}

func (u SomaUtil) TryGetTemplatePropertyByUUIDOrName(c *resty.Client, s string) string {
	if u.IsUUID(s) {
		return s
	}
	return u.GetPropertyIdByName(c, "template", s, "none")
}

func (u SomaUtil) GetPropertyIdByName(c *resty.Client, pType string, prop string, ctx string) string {
	var (
		req         proto.Request
		ctxIdString string
		path        string
	)
	req = proto.Request{
		Filter: &proto.Filter{
			Property: &proto.PropertyFilter{
				Type: pType,
				Name: prop,
			},
		},
	}

	switch pType {
	case "custom":
		// context ctx is repository
		ctxIdString = u.TryGetRepositoryByUUIDOrName(c, ctx)
		path = fmt.Sprintf("/filter/property/custom/%s/", ctxIdString)
		req.Filter.Property.RepositoryId = ctxIdString
	case "system":
		path = "/filter/property/system/"
	case "template":
		path = "/filter/property/service/global/"
	case "service":
		// context ctx is team
		ctxIdString = u.TryGetTeamByUUIDOrName(c, ctx)
		path = fmt.Sprintf("/filter/property/service/team/%s/", ctxIdString)
	default:
		u.Abort("Unsupported property type in util.GetPropertyIdByName()")
	}

	resp := u.PostRequestWithBody(c, req, path)
	res := u.DecodeProtoResultPropertyFromResponse(resp)

	if res.Properties == nil || *res.Properties == nil {
		u.Abort("Property lookup result contained no properties")
	}
	if len(*res.Properties) != 1 {
		u.Abort(fmt.Sprintf("Property lookup expected 1 result, received: %d",
			len(*res.Properties)))
	}

	switch pType {
	case "custom":
		if prop == (*res.Properties)[0].Custom.Name &&
			ctxIdString == (*res.Properties)[0].Custom.RepositoryId {
			return (*res.Properties)[0].Custom.Id
		}
	case "service":
		if ctxIdString != (*res.Properties)[0].Service.TeamId {
			goto fail
		}
		fallthrough
	case "template":
		if prop == (*res.Properties)[0].Service.Name {
			return (*res.Properties)[0].Service.Name
		}
		goto fail
	case "system":
		if prop == (*res.Properties)[0].System.Name {
			return (*res.Properties)[0].System.Name
		}
		goto fail
	}

fail:
	u.Abort("Received result set for incorrect property")

	// required to silence the compiler, since ending in a switch is not
	// analyzed to always return:
	// http://code.google.com/p/go/issues/detail?id=65
	panic("unreachable")
}

func (u SomaUtil) CheckStringIsSystemProperty(c *resty.Client, s string) {
	resp := u.GetRequest(c, "/property/system/")
	res := u.DecodeProtoResultPropertyFromResponse(resp)

	for _, prop := range *res.Properties {
		if prop.System.Name == s {
			return
		}
	}
	u.Abort(fmt.Sprintf("Invalid system property requested: %s", s))
}

func (u SomaUtil) DecodeProtoResultPropertyFromResponse(resp *resty.Response) *proto.Result {
	return u.DecodeResultFromResponse(resp)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix