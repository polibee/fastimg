package controllers

import "testing"

func TestPublicStatusNeverIncludesSecrets(t *testing.T) {
	components := publicComponents([]componentProbe{
		{Name: "api", Status: "operational", Detail: "ok"},
		{Name: "database", Status: "degraded", Detail: "password=secret"},
	})
	if components[1]["detail"] != "database is degraded" {
		t.Fatalf("status detail leaked internal message: %#v", components[1])
	}
}
