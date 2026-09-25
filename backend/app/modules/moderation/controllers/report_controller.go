package controllers

import httpcontract "github.com/goravel/framework/contracts/http"

type ReportController struct{}

func NewReportController() *ReportController { return &ReportController{} }
func (c *ReportController) Create(x httpcontract.Context) httpcontract.Response {
	return createReport(x)
}
func (c *ReportController) Mine(x httpcontract.Context) httpcontract.Response { return listReports(x) }
