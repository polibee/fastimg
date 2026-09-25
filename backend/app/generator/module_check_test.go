package generator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckModuleReportsMissingFilesWithoutWriting(t *testing.T) {
	root := t.TempDir()
	report, err := CheckModule(root, "billing")
	if err != nil {
		t.Fatalf("check module: %v", err)
	}
	if report.Complete {
		t.Fatal("expected incomplete module report")
	}
	if len(report.Missing) == 0 {
		t.Fatal("expected missing files")
	}
	entries, err := os.ReadDir(filepath.Join(root, "app", "modules"))
	if err == nil || entries != nil {
		t.Fatalf("check command wrote files: entries=%v err=%v", entries, err)
	}
}

func TestCheckModuleReportsCompleteGeneratedModule(t *testing.T) {
	root := t.TempDir()
	spec, err := Normalize(Input{Name: "billing", Fields: []string{"name:text"}})
	if err != nil {
		t.Fatal(err)
	}
	artifacts, err := RenderResourcePipeline(spec, "00000000000000")
	if err != nil {
		t.Fatal(err)
	}
	if err := WriteAll(root, artifacts); err != nil {
		t.Fatal(err)
	}
	runtimeArtifacts, err := RenderRuntimeRegistration(root, spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := WriteAll(root, runtimeArtifacts); err != nil {
		t.Fatal(err)
	}
	report, err := CheckModule(root, "billing")
	if err != nil {
		t.Fatalf("check module: %v", err)
	}
	if !report.Complete || len(report.Missing) != 0 {
		t.Fatalf("report = %+v", report)
	}
}

func TestCheckModuleRequiresRuntimeDiscoveryFiles(t *testing.T) {
	root := t.TempDir()
	spec, err := Normalize(Input{Name: "billing", Fields: []string{"name:text"}})
	if err != nil {
		t.Fatal(err)
	}
	artifacts, err := RenderResourcePipeline(spec, "00000000000000")
	if err != nil {
		t.Fatal(err)
	}
	if err := WriteAll(root, artifacts); err != nil {
		t.Fatal(err)
	}
	report, err := CheckModule(root, "billing")
	if err != nil {
		t.Fatalf("check module: %v", err)
	}
	if report.Complete || !containsPath(report.Missing, "app/modules/admin/registry/generated_resources.go") {
		t.Fatalf("report = %+v, want missing backend discovery", report)
	}
}

func TestCheckModuleSupportsFrontendSiblingWhenRunFromBackend(t *testing.T) {
	projectRoot := t.TempDir()
	backendRoot := filepath.Join(projectRoot, "backend")
	adminRoot := filepath.Join(projectRoot, "admin")
	if err := os.MkdirAll(backendRoot, 0o755); err != nil {
		t.Fatalf("create backend root: %v", err)
	}
	if err := os.MkdirAll(adminRoot, 0o755); err != nil {
		t.Fatalf("create admin root: %v", err)
	}
	if _, err := GenerateResource(backendRoot, Input{Name: "billing", Fields: []string{"name:text"}}, "20260923000000"); err != nil {
		t.Fatalf("generate resource: %v", err)
	}

	report, err := CheckModule(backendRoot, "billing")
	if err != nil {
		t.Fatalf("check module: %v", err)
	}
	if !report.Complete || len(report.Missing) != 0 {
		t.Fatalf("report = %+v, want complete sibling frontend module", report)
	}
}

func containsPath(paths []string, want string) bool {
	for _, path := range paths {
		if path == want {
			return true
		}
	}
	return false
}
