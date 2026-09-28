package backups

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	ErrDatabaseToolUnavailable     = errors.New("database backup tool is unavailable")
	ErrRestoreModeNotAllowed       = errors.New("backup restore mode is not allowed")
	ErrRestoreConfirmationRequired = errors.New("backup restore confirmation is required")
	ErrRestoreTargetNotEmpty       = errors.New("backup restore target is not empty")
)

type databaseDumpConfig struct {
	Host     string
	Port     string
	Database string
	Username string
	Password string
}

// DatabaseTool keeps pg_dump/pg_restore behind a small testable boundary. The
// command is executed directly; request values are never interpolated into a
// shell command.
type DatabaseTool interface {
	Dump(ctx context.Context, destination string) error
	Restore(ctx context.Context, source string) error
}

type commandDatabaseTool struct {
	config  databaseDumpConfig
	dump    string
	restore string
}

func NewCommandDatabaseTool() DatabaseTool {
	return &commandDatabaseTool{
		config: databaseDumpConfig{
			Host: os.Getenv("DB_HOST"), Port: os.Getenv("DB_PORT"), Database: os.Getenv("DB_DATABASE"),
			Username: os.Getenv("DB_USERNAME"), Password: os.Getenv("DB_PASSWORD"),
		},
		dump: os.Getenv("FASTIMG_PG_DUMP_PATH"), restore: os.Getenv("FASTIMG_PG_RESTORE_PATH"),
	}
}

func databaseDumpArguments(config databaseDumpConfig) []string {
	args := []string{"--format=custom", "--no-owner", "--no-privileges"}
	if config.Host != "" {
		args = append(args, "--host", config.Host)
	}
	if config.Port != "" {
		args = append(args, "--port", config.Port)
	}
	if config.Username != "" {
		args = append(args, "--username", config.Username)
	}
	args = append(args, "--file")
	return args
}

func databaseRestoreArguments(config databaseDumpConfig) []string {
	args := []string{"--clean", "--if-exists", "--no-owner", "--no-privileges"}
	if config.Host != "" {
		args = append(args, "--host", config.Host)
	}
	if config.Port != "" {
		args = append(args, "--port", config.Port)
	}
	if config.Username != "" {
		args = append(args, "--username", config.Username)
	}
	return args
}

func (t *commandDatabaseTool) Dump(ctx context.Context, destination string) error {
	if filepath.Clean(destination) != destination || destination == "" {
		return ErrDatabaseToolUnavailable
	}
	executable, err := resolveTool(t.dump, "pg_dump")
	if err != nil {
		return err
	}
	args := append(databaseDumpArguments(t.config), destination)
	if t.config.Database == "" {
		return ErrDatabaseToolUnavailable
	}
	args = append(args, t.config.Database)
	return runDatabaseCommand(ctx, executable, args, t.config.Password)
}

func (t *commandDatabaseTool) Restore(ctx context.Context, source string) error {
	if filepath.Clean(source) != source || source == "" {
		return ErrDatabaseToolUnavailable
	}
	executable, err := resolveTool(t.restore, "pg_restore")
	if err != nil {
		return err
	}
	args := databaseRestoreArguments(t.config)
	args = append(args, source)
	if t.config.Database == "" {
		return ErrDatabaseToolUnavailable
	}
	args = append(args, "--dbname", t.config.Database)
	return runDatabaseCommand(ctx, executable, args, t.config.Password)
}

func resolveTool(configured, fallback string) (string, error) {
	value := strings.TrimSpace(configured)
	if value == "" {
		value = fallback
	}
	base := strings.ToLower(filepath.Base(value))
	if base != fallback && base != fallback+".exe" {
		return "", ErrDatabaseToolUnavailable
	}
	if filepath.IsAbs(value) {
		if _, err := os.Stat(value); err != nil {
			return "", ErrDatabaseToolUnavailable
		}
		return value, nil
	}
	resolved, err := exec.LookPath(value)
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrDatabaseToolUnavailable, base)
	}
	return resolved, nil
}

func runDatabaseCommand(ctx context.Context, executable string, args []string, password string) error {
	command := exec.CommandContext(ctx, executable, args...)
	// PGPASSWORD is inherited only by this child process and is never placed in
	// argv, audit metadata, or an error message.
	if password != "" {
		command.Env = append(os.Environ(), "PGPASSWORD="+password)
	}
	if err := command.Run(); err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return fmt.Errorf("%w: timeout", ErrDatabaseToolUnavailable)
		}
		return fmt.Errorf("%w: command failed", ErrDatabaseToolUnavailable)
	}
	return nil
}

func validateRestoreRequest(request RestoreRequest) error {
	if strings.TrimSpace(request.Mode) != "new_server" {
		return ErrRestoreModeNotAllowed
	}
	if strings.TrimSpace(request.Confirmation) != "RESTORE_FASTIMG_BACKUP" {
		return ErrRestoreConfirmationRequired
	}
	return nil
}
