package backups

import (
	"archive/tar"
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/klauspost/compress/zstd"
)

var (
	ErrArchiveTooLarge       = errors.New("backup archive exceeds configured limit")
	ErrArchiveUnsafeEntry    = errors.New("backup archive contains an unsafe entry")
	ErrArchiveDuplicateEntry = errors.New("backup archive contains a duplicate entry")
	ErrArchiveChecksum       = errors.New("backup archive checksum mismatch")
	ErrArchiveMissingEntry   = errors.New("backup archive is missing a required entry")
)

type archiveEntry struct {
	name string
	body []byte
}

// ValidateArchive validates and reads an archive without extracting any entry
// into the live filesystem. The caller may safely use the returned preview to
// decide whether to stage a restore.
func ValidateArchive(ctx context.Context, source io.ReaderAt, size int64, limits ArchiveLimits) (ValidationPreview, error) {
	if source == nil || size <= 0 || limits.MaxArchiveBytes <= 0 || size > limits.MaxArchiveBytes {
		return ValidationPreview{}, ErrArchiveTooLarge
	}
	if limits.MaxEntries <= 0 || limits.MaxExpandedBytes <= 0 || limits.MaxFileBytes <= 0 {
		return ValidationPreview{}, ErrArchiveTooLarge
	}
	section := io.NewSectionReader(source, 0, size)
	decoder, err := zstd.NewReader(section)
	if err != nil {
		return ValidationPreview{}, fmt.Errorf("open backup archive: %w", err)
	}
	defer decoder.Close()

	reader := tar.NewReader(decoder)
	entries := make(map[string][]byte)
	var expanded int64
	var count int64
	for {
		if err := ctx.Err(); err != nil {
			return ValidationPreview{}, err
		}
		header, nextErr := reader.Next()
		if errors.Is(nextErr, io.EOF) {
			break
		}
		if nextErr != nil {
			return ValidationPreview{}, fmt.Errorf("read backup archive: %w", nextErr)
		}
		count++
		if count > limits.MaxEntries {
			return ValidationPreview{}, ErrArchiveTooLarge
		}
		if err := validateEntryName(header.Name); err != nil {
			return ValidationPreview{}, err
		}
		if header.Typeflag != tar.TypeReg && header.Typeflag != 0 {
			return ValidationPreview{}, fmt.Errorf("%w: %s", ErrArchiveUnsafeEntry, header.Name)
		}
		if header.Mode&0o111 != 0 || isExecutableName(header.Name) {
			return ValidationPreview{}, fmt.Errorf("%w: executable %s", ErrArchiveUnsafeEntry, header.Name)
		}
		if _, exists := entries[header.Name]; exists {
			return ValidationPreview{}, fmt.Errorf("%w: %s", ErrArchiveDuplicateEntry, header.Name)
		}
		if header.Size < 0 || header.Size > limits.MaxFileBytes || header.Size > limits.MaxExpandedBytes-expanded {
			return ValidationPreview{}, ErrArchiveTooLarge
		}
		body, readErr := io.ReadAll(io.LimitReader(reader, limits.MaxFileBytes+1))
		if readErr != nil {
			return ValidationPreview{}, fmt.Errorf("read backup entry %s: %w", header.Name, readErr)
		}
		if int64(len(body)) != header.Size {
			return ValidationPreview{}, fmt.Errorf("%w: truncated %s", ErrArchiveUnsafeEntry, header.Name)
		}
		expanded += int64(len(body))
		entries[header.Name] = body
	}

	for _, required := range []string{"manifest.json", "database.dump", "settings.json", "checksums.sha256"} {
		if _, ok := entries[required]; !ok {
			return ValidationPreview{}, fmt.Errorf("%w: %s", ErrArchiveMissingEntry, required)
		}
	}
	manifest, err := parseManifest(entries["manifest.json"])
	if err != nil {
		return ValidationPreview{}, err
	}
	if err := verifyChecksums(entries["checksums.sha256"], entries); err != nil {
		return ValidationPreview{}, err
	}
	files := make([]string, 0, len(entries))
	for name := range entries {
		files = append(files, name)
	}
	sort.Strings(files)
	return ValidationPreview{
		FormatVersion:    int64(manifest.FormatVersion),
		FileCount:        count,
		ExpandedBytes:    expanded,
		Files:            files,
		SettingsExcluded: append([]string(nil), manifest.SettingsExcluded...),
	}, nil
}

func validateEntryName(name string) error {
	if name == "" || strings.Contains(name, "\\") || strings.HasPrefix(name, "/") || strings.HasPrefix(name, "//") || (len(name) >= 2 && name[1] == ':') || path.Clean(name) != name {
		return fmt.Errorf("%w: %q", ErrArchiveUnsafeEntry, name)
	}
	for _, segment := range strings.Split(name, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return fmt.Errorf("%w: %q", ErrArchiveUnsafeEntry, name)
		}
	}
	if name != "manifest.json" && name != "database.dump" && name != "settings.json" && name != "checksums.sha256" && !strings.HasPrefix(name, "storage/media/") && !strings.HasPrefix(name, "storage/variants/") {
		return fmt.Errorf("%w: unsupported entry %q", ErrArchiveUnsafeEntry, name)
	}
	return nil
}

func isExecutableName(name string) bool {
	switch strings.ToLower(path.Ext(name)) {
	case ".exe", ".com", ".bat", ".cmd", ".dll", ".sh", ".ps1", ".msi":
		return true
	default:
		return false
	}
}

func parseManifest(body []byte) (Manifest, error) {
	var manifest Manifest
	if err := json.Unmarshal(body, &manifest); err != nil {
		return Manifest{}, fmt.Errorf("invalid backup manifest: %w", err)
	}
	if manifest.FormatVersion != archiveFormatVersion {
		return Manifest{}, fmt.Errorf("unsupported backup format version %d", manifest.FormatVersion)
	}
	return manifest, nil
}

func verifyChecksums(body []byte, entries map[string][]byte) error {
	seen := make(map[string]struct{})
	scanner := bufio.NewScanner(strings.NewReader(string(body)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) != 2 || len(parts[0]) != sha256.Size*2 {
			return ErrArchiveChecksum
		}
		if _, err := hex.DecodeString(parts[0]); err != nil {
			return ErrArchiveChecksum
		}
		if err := validateEntryName(parts[1]); err != nil || parts[1] == "checksums.sha256" {
			return ErrArchiveChecksum
		}
		if _, duplicate := seen[parts[1]]; duplicate {
			return ErrArchiveChecksum
		}
		content, ok := entries[parts[1]]
		if !ok {
			return ErrArchiveChecksum
		}
		digest := sha256.Sum256(content)
		if !strings.EqualFold(parts[0], hex.EncodeToString(digest[:])) {
			return ErrArchiveChecksum
		}
		seen[parts[1]] = struct{}{}
	}
	if err := scanner.Err(); err != nil {
		return ErrArchiveChecksum
	}
	for name := range entries {
		if name != "checksums.sha256" {
			if _, ok := seen[name]; !ok {
				return ErrArchiveChecksum
			}
		}
	}
	return nil
}

type ArchiveFile struct {
	Name string
	Body []byte
}

// WriteArchive creates a portable archive from already-authorized files. The
// service layer is responsible for obtaining database/media bytes through its
// own boundaries; this function only writes the safe archive format.
func WriteArchive(ctx context.Context, destination string, jobID uint64, files []ArchiveFile, excludedSecrets []string) (Manifest, error) {
	if strings.TrimSpace(destination) == "" {
		return Manifest{}, errors.New("backup destination is required")
	}
	contents := make(map[string][]byte, len(files)+3)
	for _, file := range files {
		if err := validateEntryName(file.Name); err != nil || file.Name == "checksums.sha256" || file.Name == "manifest.json" {
			return Manifest{}, fmt.Errorf("invalid backup file %q", file.Name)
		}
		if file.Name == "database.dump" || file.Name == "settings.json" || strings.HasPrefix(file.Name, "storage/") {
			contents[file.Name] = append([]byte(nil), file.Body...)
		} else {
			return Manifest{}, fmt.Errorf("invalid backup file %q", file.Name)
		}
	}
	manifest := Manifest{FormatVersion: archiveFormatVersion, JobID: jobID, CreatedAt: time.Now().UTC(), SettingsExcluded: append([]string(nil), excludedSecrets...)}
	for name := range contents {
		manifest.Files = append(manifest.Files, name)
	}
	sort.Strings(manifest.Files)
	manifestBody, err := json.Marshal(manifest)
	if err != nil {
		return Manifest{}, err
	}
	contents["manifest.json"] = manifestBody
	var checksumLines []string
	for _, name := range append([]string{"manifest.json"}, manifest.Files...) {
		digest := sha256.Sum256(contents[name])
		checksumLines = append(checksumLines, fmt.Sprintf("%s  %s", hex.EncodeToString(digest[:]), name))
	}
	contents["checksums.sha256"] = []byte(strings.Join(checksumLines, "\n") + "\n")

	file, err := os.Create(destination)
	if err != nil {
		return Manifest{}, err
	}
	defer file.Close()
	encoder, err := zstd.NewWriter(file)
	if err != nil {
		return Manifest{}, err
	}
	writer := tar.NewWriter(encoder)
	archiveNames := append([]string{"manifest.json"}, manifest.Files...)
	archiveNames = append(archiveNames, "checksums.sha256")
	for _, name := range archiveNames {
		if err := ctx.Err(); err != nil {
			_ = writer.Close()
			_ = encoder.Close()
			return Manifest{}, err
		}
		body := contents[name]
		if err := writer.WriteHeader(&tar.Header{Name: name, Mode: 0o600, Size: int64(len(body)), Typeflag: tar.TypeReg}); err != nil {
			return Manifest{}, err
		}
		if _, err := writer.Write(body); err != nil {
			return Manifest{}, err
		}
	}
	if err := writer.Close(); err != nil {
		return Manifest{}, err
	}
	if err := encoder.Close(); err != nil {
		return Manifest{}, err
	}
	return manifest, nil
}
