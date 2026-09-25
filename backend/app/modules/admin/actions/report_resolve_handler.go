package actions

import (
	"errors"
	h "github.com/goravel/framework/contracts/http"
	a "goravel/app/core/admin/actions"
	"goravel/app/facades"
	"strconv"
)

type ResolveHandler struct{}

func NewResolveHandler() *ResolveHandler { return &ResolveHandler{} }
func (*ResolveHandler) Kind() string     { return "report-resolve" }
func (*ResolveHandler) Payload() string  { return "report-resolve" }
func (hh *ResolveHandler) Execute(ctx h.Context, r a.Request) (a.Result, error) {
	act, n, e := ParseResolvePayload(r.Payload)
	if e != nil {
		return a.Result{}, e
	}
	v, e := facades.Auth(ctx).ID()
	if e != nil {
		return a.Result{}, e
	}
	op, e := strconv.ParseUint(v, 10, 32)
	if e != nil || op == 0 {
		return a.Result{}, errors.New("invalid operator")
	}
	out := a.Result{Action: r.Action, Requested: len(r.IDs)}
	for _, id := range r.IDs {
		if e := resolveReport(uint(id), uint(op), act, n); e != nil {
			out.Failed++
			out.Failures = append(out.Failures, a.Failure{ID: id, Code: "REPORT_RESOLVE_FAILED"})
		} else {
			out.Succeeded++
		}
	}
	return out, nil
}
