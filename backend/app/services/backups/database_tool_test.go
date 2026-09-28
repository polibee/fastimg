package backups

import (
	"reflect"
	"testing"
)

func TestDatabaseToolArgumentsAreStructuredAndDoNotContainSecrets(t *testing.T) {
	args := databaseDumpArguments(databaseDumpConfig{
		Host: "127.0.0.1", Port: "5432", Database: "fastimg", Username: "fastimg",
	})
	want := []string{"--format=custom", "--no-owner", "--no-privileges", "--host", "127.0.0.1", "--port", "5432", "--username", "fastimg", "--file"}
	if !reflect.DeepEqual(args[:len(want)], want) {
		t.Fatalf("databaseDumpArguments() prefix = %#v, want %#v", args[:len(want)], want)
	}
	for _, arg := range args {
		if arg == "fastimg-password" || arg == "DB_PASSWORD" {
			t.Fatalf("database password leaked into argv: %#v", args)
		}
	}
}

func TestRestoreRequestRequiresNewServerConfirmation(t *testing.T) {
	if err := validateRestoreRequest(RestoreRequest{Mode: "existing", Confirmation: "RESTORE_FASTIMG_BACKUP"}); err != ErrRestoreModeNotAllowed {
		t.Fatalf("mode error = %v", err)
	}
	if err := validateRestoreRequest(RestoreRequest{Mode: "new_server"}); err != ErrRestoreConfirmationRequired {
		t.Fatalf("confirmation error = %v", err)
	}
}
