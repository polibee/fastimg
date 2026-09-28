package backups

import "time"

const archiveFormatVersion = 1

// Manifest describes the portable, non-secret part of a FastImg backup.
// Secret values are represented only by their keys in SettingsExcluded.
type Manifest struct {
	FormatVersion    int       `json:"format_version"`
	JobID            uint64    `json:"job_id"`
	CreatedAt        time.Time `json:"created_at"`
	Files            []string  `json:"files"`
	SettingsExcluded []string  `json:"settings_excluded"`
}

type ValidationPreview struct {
	FormatVersion    int64    `json:"format_version"`
	FileCount        int64    `json:"file_count"`
	ExpandedBytes    int64    `json:"expanded_bytes"`
	Files            []string `json:"files"`
	SettingsExcluded []string `json:"settings_excluded"`
}

type ArchiveLimits struct {
	MaxArchiveBytes  int64
	MaxEntries       int64
	MaxExpandedBytes int64
	MaxFileBytes     int64
}

func DefaultArchiveLimits() ArchiveLimits {
	return ArchiveLimits{
		MaxArchiveBytes:  10 << 30,
		MaxEntries:       200000,
		MaxExpandedBytes: 100 << 30,
		MaxFileBytes:     10 << 30,
	}
}
