package migrations

import "goravel/app/facades"

// M20260928000004AddBackupActiveJobIndex prevents two processes from
// assembling a large backup or restore archive at the same time.
type M20260928000004AddBackupActiveJobIndex struct{}

func (m *M20260928000004AddBackupActiveJobIndex) Signature() string {
	return "20260928000004_add_backup_active_job_index"
}

func (m *M20260928000004AddBackupActiveJobIndex) Up() error {
	if !facades.Schema().HasTable("backup_jobs") {
		return nil
	}
	_, err := facades.Orm().Query().Exec("CREATE UNIQUE INDEX IF NOT EXISTS backup_jobs_single_active ON backup_jobs ((1)) WHERE status IN ('queued', 'running', 'restoring')")
	return err
}

func (m *M20260928000004AddBackupActiveJobIndex) Down() error {
	_, err := facades.Orm().Query().Exec("DROP INDEX IF EXISTS backup_jobs_single_active")
	return err
}
