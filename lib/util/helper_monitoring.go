package util

import (
	"fmt"

	"github.com/1and1/soma/lib/adm"
	"github.com/1and1/soma/lib/proto"
	"gopkg.in/resty.v0"
)

func (u *SomaUtil) TryGetMonitoringByUUIDOrName(c *resty.Client, s string) string {
	if u.IsUUID(s) {
		return s
	}
	return u.GetMonitoringIdByName(c, s)
}

func (u *SomaUtil) GetMonitoringIdByName(c *resty.Client, monitoring string) string {
	req := proto.NewMonitoringFilter()
	req.Filter.Monitoring.Name = monitoring

	resp, err := adm.PostReqBody(req, `/filter/monitoring/`)
	if err != nil {
		u.Abort(fmt.Sprintf("Monitoring lookup request error: %s", err.Error()))
	}
	result, err := u.ResultFromResponse(resp)
	if se, ok := err.(SomaError); ok {
		if se.RequestError() {
			u.Abort(fmt.Sprintf("Monitoring lookup request error: %s", se.Error()))
		}
		if se.Code() == 404 {
			u.Abort(fmt.Sprintf(
				"Could not find monitoring system with name %s",
				monitoring,
			))
		}
		u.Abort(fmt.Sprintf("Monitoring lookup application error: %s", err.Error()))
	}

	if monitoring != (*result.Monitorings)[0].Name {
		u.Abort(fmt.Sprintf(
			"Monitoring system ID lookup failed. Wanted %s, received %s",
			monitoring,
			(*result.Monitorings)[0].Name,
		))
	}
	return (*result.Monitorings)[0].Id
}

func (u *SomaUtil) DecodeProtoResultMonitoringFromResponse(resp *resty.Response) *proto.Result {
	return u.DecodeResultFromResponse(resp)
}

// vim: ts=4 sw=4 sts=4 noet fenc=utf-8 ffs=unix