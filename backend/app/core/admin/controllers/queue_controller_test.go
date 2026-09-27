package controllers

import "testing"

func TestQueueTaskPermissionsAreExplicit(t *testing.T) {
	if QueueViewPermission != "admin.tasks.view" {
		t.Fatalf("view permission = %q", QueueViewPermission)
	}
	if QueueRetryPermission != "admin.tasks.retry" {
		t.Fatalf("retry permission = %q", QueueRetryPermission)
	}
}

func TestQueueTaskPageMetadataIsStable(t *testing.T) {
	items := []queueTaskItem{{UUID: "one"}, {UUID: "two"}, {UUID: "three"}}
	page := paginateQueueTasks(items, 2, 2)
	if page.Meta.Page != 2 || page.Meta.PerPage != 2 || page.Meta.Total != 3 || page.Meta.LastPage != 2 {
		t.Fatalf("unexpected meta: %+v", page.Meta)
	}
	if len(page.Data) != 1 || page.Data[0].UUID != "three" {
		t.Fatalf("unexpected page data: %+v", page.Data)
	}
}
