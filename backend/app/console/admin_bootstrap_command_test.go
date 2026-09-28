package console

import "testing"

func TestAdminBootstrapCommandMetadata(t *testing.T) {
	command := AdminBootstrapCommand{}
	if command.Signature() != "admin:bootstrap" {
		t.Fatalf("signature = %q", command.Signature())
	}
	if command.Description() == "" {
		t.Fatal("description must not be empty")
	}
}
