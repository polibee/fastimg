package controllers

import "testing"

func TestBackupPermissionContract(t *testing.T) {
	if BackupManagePermission != "admin.backups.manage" {
		t.Fatalf("manage permission = %q", BackupManagePermission)
	}
	if BackupDownloadPermission != "admin.backups.download" {
		t.Fatalf("download permission = %q", BackupDownloadPermission)
	}
}

