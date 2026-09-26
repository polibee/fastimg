package backups

import (
	"archive/tar"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/klauspost/compress/zstd"
	"github.com/stretchr/testify/require"
)

type testArchiveEntry struct {
	name     string
	body     string
	typeFlag byte
	linkname string
	mode     int64
}

func makeTestArchive(t *testing.T, entries ...testArchiveEntry) []byte {
	t.Helper()
	var compressed bytes.Buffer
	encoder, err := zstd.NewWriter(&compressed)
	require.NoError(t, err)
	writer := tar.NewWriter(encoder)
	for _, entry := range entries {
		header := &tar.Header{Name: entry.name, Mode: entry.mode, Size: int64(len(entry.body)), Typeflag: entry.typeFlag, Linkname: entry.linkname}
		if header.Mode == 0 {
			header.Mode = 0o600
		}
		require.NoError(t, writer.WriteHeader(header))
		if entry.typeFlag == tar.TypeReg || entry.typeFlag == 0 {
			_, err = io.WriteString(writer, entry.body)
			require.NoError(t, err)
		}
	}
	require.NoError(t, writer.Close())
	require.NoError(t, encoder.Close())
	return compressed.Bytes()
}

func validTestArchive(t *testing.T) []byte {
	t.Helper()
	files := map[string]string{
		"manifest.json": `{"format_version":1}`,
		"database.dump": "database",
		"settings.json": `{"site_title":"FastImg"}`,
	}
	var checksumLines []string
	entries := make([]testArchiveEntry, 0, len(files)+1)
	for name, body := range files {
		digest := sha256.Sum256([]byte(body))
		checksumLines = append(checksumLines, fmt.Sprintf("%s  %s", hex.EncodeToString(digest[:]), name))
		entries = append(entries, testArchiveEntry{name: name, body: body, typeFlag: tar.TypeReg})
	}
	entries = append(entries, testArchiveEntry{name: "checksums.sha256", body: strings.Join(checksumLines, "\n") + "\n", typeFlag: tar.TypeReg})
	return makeTestArchive(t, entries...)
}

func TestValidateArchiveRejectsUnsafePathsAndLinks(t *testing.T) {
	tests := []struct {
		name  string
		entry testArchiveEntry
	}{
		{name: "parent traversal", entry: testArchiveEntry{name: "storage/media/../outside.txt", body: "x", typeFlag: tar.TypeReg}},
		{name: "absolute path", entry: testArchiveEntry{name: "/tmp/outside.txt", body: "x", typeFlag: tar.TypeReg}},
		{name: "windows drive path", entry: testArchiveEntry{name: "C:/outside.txt", body: "x", typeFlag: tar.TypeReg}},
		{name: "backslash path", entry: testArchiveEntry{name: "storage\\media\\outside.txt", body: "x", typeFlag: tar.TypeReg}},
		{name: "symlink", entry: testArchiveEntry{name: "storage/media/link", typeFlag: tar.TypeSymlink, linkname: "../../outside"}},
		{name: "hardlink", entry: testArchiveEntry{name: "storage/media/link", typeFlag: tar.TypeLink, linkname: "storage/media/file"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			archive := makeTestArchive(t, tt.entry)
			_, err := ValidateArchive(contextWithTestArchive(t), bytes.NewReader(archive), int64(len(archive)), DefaultArchiveLimits())
			require.Error(t, err)
		})
	}
}

func TestValidateArchiveRejectsDuplicatesExecutablesAndLimits(t *testing.T) {
	tests := []struct {
		name    string
		limits  ArchiveLimits
		entries []testArchiveEntry
	}{
		{name: "duplicate", entries: []testArchiveEntry{{name: "manifest.json", body: "{}", typeFlag: tar.TypeReg}, {name: "manifest.json", body: "{}", typeFlag: tar.TypeReg}}},
		{name: "executable extension", entries: []testArchiveEntry{{name: "storage/media/run.exe", body: "MZ", typeFlag: tar.TypeReg}}},
		{name: "executable mode", entries: []testArchiveEntry{{name: "storage/media/run.bin", body: "x", typeFlag: tar.TypeReg, mode: 0o700}}},
		{name: "entry limit", limits: ArchiveLimits{MaxEntries: 1, MaxExpandedBytes: 1024, MaxFileBytes: 1024}, entries: []testArchiveEntry{{name: "manifest.json", body: "{}", typeFlag: tar.TypeReg}, {name: "database.dump", body: "x", typeFlag: tar.TypeReg}}},
		{name: "file limit", limits: ArchiveLimits{MaxEntries: 10, MaxExpandedBytes: 1024, MaxFileBytes: 1}, entries: []testArchiveEntry{{name: "manifest.json", body: "{}", typeFlag: tar.TypeReg}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limits := tt.limits
			if limits.MaxEntries == 0 {
				limits = DefaultArchiveLimits()
			}
			archive := makeTestArchive(t, tt.entries...)
			_, err := ValidateArchive(contextWithTestArchive(t), bytes.NewReader(archive), int64(len(archive)), limits)
			require.Error(t, err)
		})
	}
}

func TestValidateArchiveRejectsChecksumMismatch(t *testing.T) {
	archive := makeTestArchive(t,
		testArchiveEntry{name: "manifest.json", body: `{}`, typeFlag: tar.TypeReg},
		testArchiveEntry{name: "database.dump", body: "database", typeFlag: tar.TypeReg},
		testArchiveEntry{name: "settings.json", body: `{}`, typeFlag: tar.TypeReg},
		testArchiveEntry{name: "checksums.sha256", body: "0000000000000000000000000000000000000000000000000000000000000000  manifest.json\n", typeFlag: tar.TypeReg},
	)
	_, err := ValidateArchive(contextWithTestArchive(t), bytes.NewReader(archive), int64(len(archive)), DefaultArchiveLimits())
	require.Error(t, err)
}

func TestValidateArchiveReturnsPreviewForValidArchive(t *testing.T) {
	archive := validTestArchive(t)
	preview, err := ValidateArchive(contextWithTestArchive(t), bytes.NewReader(archive), int64(len(archive)), DefaultArchiveLimits())
	require.NoError(t, err)
	require.Equal(t, int64(4), preview.FileCount)
	require.Greater(t, preview.ExpandedBytes, int64(0))
	require.Contains(t, preview.Files, "database.dump")
}

func TestWriteArchiveRoundTripsThroughValidator(t *testing.T) {
	destination := t.TempDir() + "\\backup.tar.zst"
	manifest, err := WriteArchive(context.Background(), destination, 42, []ArchiveFile{
		{Name: "database.dump", Body: []byte("database")},
		{Name: "settings.json", Body: []byte(`{"site_title":"FastImg"}`)},
		{Name: "storage/media/1/original.png", Body: []byte("png")},
	}, []string{"email.smtp.password"})
	require.NoError(t, err)
	require.Equal(t, uint64(42), manifest.JobID)
	archive, err := os.Open(destination)
	require.NoError(t, err)
	defer archive.Close()
	stat, err := archive.Stat()
	require.NoError(t, err)
	preview, err := ValidateArchive(context.Background(), archive, stat.Size(), DefaultArchiveLimits())
	require.NoError(t, err)
	require.Equal(t, int64(5), preview.FileCount)
	require.Contains(t, preview.SettingsExcluded, "email.smtp.password")
}

// contextWithTestArchive keeps the tests explicit about cancellation without
// coupling the archive contract to an HTTP request context.
func contextWithTestArchive(t *testing.T) context.Context {
	t.Helper()
	return context.Background()
}
