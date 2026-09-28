package openapi

import "testing"

func TestSpecCoversImplementedAdminContracts(t *testing.T) {
	spec := Spec()
	if spec["openapi"] != "3.0.3" {
		t.Fatalf("unexpected OpenAPI version: %v", spec["openapi"])
	}
	paths := spec["paths"].(map[string]any)
	for _, path := range []string{"/auth/registration-policy", "/auth/register", "/auth/verify-email", "/auth/resend-verification", "/auth/login", "/auth/refresh", "/auth/logout-all", "/plans", "/discovery/status", "/discovery/feed", "/discovery/media/{id}/content", "/public/albums/{id}", "/public/albums/{id}/media/{media_id}/content", "/subscription", "/me/usage", "/me/usage/ledger", "/admin/registry", "/admin/search", "/admin/{resource}", "/admin/{resource}/export", "/admin/{resource}/actions/{action}", "/admin/{resource}/relations/{relation}/options", "/admin/{resource}/{id}/relations/{relation}", "/admin/{resource}/{id}", "/admin/overview", "/admin/audit-logs", "/admin/audit-logs/cleanup", "/admin/media-access-logs", "/admin/tasks", "/admin/tasks/{uuid}/retry", "/admin/storage/statistics", "/notifications", "/notifications/unread-count", "/notifications/{id}/read", "/notifications/read-all", "/admin/settings", "/admin/settings/{key}", "/admin/roles/{id}/permissions"} {
		if _, ok := paths[path]; !ok {
			t.Fatalf("missing contract path %s", path)
		}
	}
	if _, ok := paths["/admin/users/status"]; ok {
		t.Fatal("legacy user status action remains in contract")
	}
	if _, ok := paths["/admin/resources/{resource}"]; ok {
		t.Fatal("legacy resources route remains in contract")
	}
	resourceCollection := paths["/admin/{resource}"].(map[string]any)
	for _, method := range []string{"get", "post"} {
		if _, ok := resourceCollection[method]; !ok {
			t.Fatalf("missing resource collection method %s", method)
		}
	}
	exportContract := paths["/admin/{resource}/export"].(map[string]any)
	exportResponse := exportContract["responses"].(map[string]any)["200"].(map[string]any)
	if _, ok := exportResponse["content"].(map[string]any)["text/csv"]; !ok {
		t.Fatal("resource export must return text/csv")
	}
	resourceItem := paths["/admin/{resource}/{id}"].(map[string]any)
	for _, method := range []string{"get", "put", "delete"} {
		if _, ok := resourceItem[method]; !ok {
			t.Fatalf("missing resource item method %s", method)
		}
	}
	schemas := spec["components"].(map[string]any)["schemas"].(map[string]any)
	for _, schema := range []string{"ResourceManifest", "ResourceField", "ResourceOption", "ResourceAction", "ActionPayloadField", "ResourceRelation", "ResourceFormGroup", "ResourceDetailSection", "ResourceFieldDependency", "RelationOption", "RelationOptionList", "ResourceListMeta", "ActionRequest", "ActionResponse", "GlobalSearchResult", "GlobalSearchResponse", "DataScope", "AuditCleanupMode", "AuditCleanupRequest", "AuditCleanupResponse", "Notification", "NotificationReadResponse", "NotificationMarkAllReadResponse", "NotificationUnreadCount", "DiscoveryStatus", "DiscoveryFeedItem", "DiscoveryFeedResponse", "StorageStatisticsOverview"} {
		if _, ok := schemas[schema]; !ok {
			t.Fatalf("missing resource schema %s", schema)
		}
	}
	cleanup := paths["/admin/audit-logs/cleanup"].(map[string]any)["post"].(map[string]any)
	cleanupBody := cleanup["requestBody"].(map[string]any)["content"].(map[string]any)["application/json"].(map[string]any)["schema"].(map[string]any)
	if cleanupBody["$ref"] != "#/components/schemas/AuditCleanupRequest" {
		t.Fatalf("unexpected audit cleanup request schema: %v", cleanupBody)
	}
	cleanupProperties := schemas["AuditCleanupRequest"].(map[string]any)["properties"].(map[string]any)
	if _, ok := cleanupProperties["mode"]; !ok {
		t.Fatal("audit cleanup contract must expose mode")
	}
	for _, property := range []string{"ids", "action", "user_id", "confirmation"} {
		if _, ok := cleanupProperties[property]; !ok {
			t.Fatalf("audit cleanup contract must expose %s", property)
		}
	}
	actionContract := paths["/admin/{resource}/actions/{action}"].(map[string]any)["post"].(map[string]any)
	if actionContract["operationId"] != "executeResourceAction" {
		t.Fatalf("unexpected resource action operation: %v", actionContract["operationId"])
	}
	status := spec["components"].(map[string]any)["schemas"].(map[string]any)["UserStatus"]
	if status == nil {
		t.Fatal("missing UserStatus schema")
	}
	rolePermissions := paths["/admin/roles/{id}/permissions"].(map[string]any)["put"].(map[string]any)
	requestBody := rolePermissions["requestBody"].(map[string]any)
	content := requestBody["content"].(map[string]any)["application/json"].(map[string]any)
	bodySchema := content["schema"].(map[string]any)
	properties := bodySchema["properties"].(map[string]any)
	if _, ok := properties["scopes"]; !ok {
		t.Fatal("role permission contract must expose scopes")
	}
	if _, ok := properties["fields"]; !ok {
		t.Fatal("role permission contract must expose fields")
	}
}

func TestSpecDocumentsPublicAndAdminContentPageContracts(t *testing.T) {
	spec := Spec()
	paths := spec["paths"].(map[string]any)
	for _, path := range []string{"/site/pages/{slug}", "/admin/content-pages", "/admin/content-pages/{id}", "/admin/content-pages/{id}/actions/publish", "/admin/content-pages/{id}/actions/archive"} {
		if _, ok := paths[path]; !ok {
			t.Fatalf("missing content page path %s", path)
		}
	}
	schemas := spec["components"].(map[string]any)["schemas"].(map[string]any)
	for _, schema := range []string{"PublicSitePage", "AdminSitePage"} {
		if _, ok := schemas[schema]; !ok {
			t.Fatalf("missing content page schema %s", schema)
		}
	}
}

func TestSpecDocumentsFooterNavigationAndFriendLinkContracts(t *testing.T) {
	spec := Spec()
	paths := spec["paths"].(map[string]any)
	for _, path := range []string{"/site/footer-navigation", "/site/friend-links/presentation", "/friend-links", "/admin/footer-navigation", "/admin/footer-navigation/groups", "/admin/footer-navigation/items", "/admin/friend-links", "/admin/friend-links/{id}/review"} {
		if _, ok := paths[path]; !ok {
			t.Fatalf("missing footer/friend-link path %s", path)
		}
	}
	schemas := spec["components"].(map[string]any)["schemas"].(map[string]any)
	for _, schema := range []string{"FooterNavigationGroup", "FooterNavigationItem", "FooterNavigationResponse", "FriendLinkPublic", "FriendLinkPresentation", "FriendLinkSubmission"} {
		if _, ok := schemas[schema]; !ok {
			t.Fatalf("missing footer/friend-link schema %s", schema)
		}
	}
}

func TestSpecDocumentsMemberAdvertisingContract(t *testing.T) {
	spec := Spec()
	paths := spec["paths"].(map[string]any)
	ads, ok := paths["/ads"].(map[string]any)
	if !ok {
		t.Fatal("missing member advertising path")
	}
	get := ads["get"].(map[string]any)
	if get["operationId"] != "listPublishedAds" {
		t.Fatalf("unexpected advertising operation id: %v", get["operationId"])
	}
	parameters := get["parameters"].([]map[string]any)
	placementSchema := parameters[0]["schema"].(map[string]any)
	enum := placementSchema["enum"].([]string)
	if len(enum) != 4 || enum[0] != "header" || enum[3] != "right" {
		t.Fatalf("unexpected advertising placement enum: %v", enum)
	}
	schemas := spec["components"].(map[string]any)["schemas"].(map[string]any)
	if _, ok := schemas["AdSlot"]; !ok {
		t.Fatal("missing AdSlot schema")
	}
}

func TestSpecDocumentsMemberTrashEmptyContract(t *testing.T) {
	spec := Spec()
	paths := spec["paths"].(map[string]any)
	trash := paths["/media/trash"].(map[string]any)["delete"].(map[string]any)
	if trash["operationId"] != "emptyOwnMediaTrash" {
		t.Fatalf("unexpected trash operation id: %v", trash["operationId"])
	}
	requestBody := trash["requestBody"].(map[string]any)["content"].(map[string]any)["application/json"].(map[string]any)["schema"].(map[string]any)
	if requestBody["type"] != "object" {
		t.Fatalf("unexpected trash request schema: %v", requestBody)
	}
	requestProperties := requestBody["properties"].(map[string]any)
	confirm := requestProperties["confirm"].(map[string]any)
	if confirm["enum"].([]string)[0] != "empty-trash" {
		t.Fatalf("unexpected trash confirmation enum: %v", confirm["enum"])
	}
	response := trash["responses"].(map[string]any)["200"].(map[string]any)["content"].(map[string]any)["application/json"].(map[string]any)["schema"].(map[string]any)
	if response["type"] != "object" {
		t.Fatalf("unexpected trash response schema: %v", response)
	}
	responseContent := trash["responses"].(map[string]any)["200"].(map[string]any)["content"].(map[string]any)["application/json"].(map[string]any)
	if responseContent["x-schema-name"] != "MediaTrashEmptyResponse" {
		t.Fatalf("unexpected trash response schema name: %v", responseContent["x-schema-name"])
	}
}

func TestSpecDocumentsFastImgUploadAndPrivateMediaRoutes(t *testing.T) {
	spec := Spec()
	paths := spec["paths"].(map[string]any)
	for _, path := range []string{"/uploads", "/uploads/batch", "/uploads/{id}", "/uploads/{id}/retry", "/media", "/media/{id}/content", "/media/{id}", "/media/{id}/restore", "/media/{id}/permanent"} {
		if _, ok := paths[path]; !ok {
			t.Fatalf("missing FastImg contract path %s", path)
		}
	}
	upload := paths["/uploads"].(map[string]any)["post"].(map[string]any)
	if upload["operationId"] != "createMediaUpload" {
		t.Fatalf("unexpected upload operation id: %v", upload["operationId"])
	}
	content := upload["requestBody"].(map[string]any)["content"].(map[string]any)
	if _, ok := content["multipart/form-data"]; !ok {
		t.Fatal("media upload must document multipart/form-data")
	}
	parameters := upload["parameters"].([]map[string]any)
	if parameters[0]["name"] != "Idempotency-Key" || parameters[0]["in"] != "header" {
		t.Fatal("media upload must require an Idempotency-Key header")
	}
	if _, ok := upload["responses"].(map[string]any)["429"]; !ok {
		t.Fatal("media upload must document quota-exceeded responses")
	}
	batchUpload := paths["/uploads/batch"].(map[string]any)["post"].(map[string]any)
	if batchUpload["operationId"] != "createBatchMediaUpload" {
		t.Fatalf("unexpected batch upload operation id: %v", batchUpload["operationId"])
	}
	if _, ok := batchUpload["responses"].(map[string]any)["207"]; !ok {
		t.Fatal("batch upload must document multi-status response")
	}
	batchBody := batchUpload["requestBody"].(map[string]any)["content"].(map[string]any)["multipart/form-data"].(map[string]any)
	batchSchema := batchBody["schema"].(map[string]any)
	filesSchema := batchSchema["properties"].(map[string]any)["files"].(map[string]any)
	if filesSchema["maxItems"] != 5 {
		t.Fatalf("unexpected batch file limit: %v", filesSchema["maxItems"])
	}
	uploadStatus := paths["/uploads/{id}"].(map[string]any)["get"].(map[string]any)
	if uploadStatus["operationId"] != "getOwnUploadStatus" {
		t.Fatalf("unexpected upload status operation id: %v", uploadStatus["operationId"])
	}
	uploadRetry := paths["/uploads/{id}/retry"].(map[string]any)["post"].(map[string]any)
	if uploadRetry["operationId"] != "retryOwnUploadSession" {
		t.Fatalf("unexpected upload retry operation id: %v", uploadRetry["operationId"])
	}
	mediaList := paths["/media"].(map[string]any)["get"].(map[string]any)
	if mediaList["operationId"] != "listOwnMedia" {
		t.Fatalf("unexpected media list operation id: %v", mediaList["operationId"])
	}
	mediaDetails := paths["/media/{id}"].(map[string]any)["get"].(map[string]any)
	if mediaDetails["operationId"] != "getOwnMediaDetails" {
		t.Fatalf("unexpected media details operation id: %v", mediaDetails["operationId"])
	}
	schemas := Spec()["components"].(map[string]any)["schemas"].(map[string]any)
	mediaItem := schemas["MediaItem"].(map[string]any)["properties"].(map[string]any)
	if _, ok := mediaItem["links"]; !ok {
		t.Fatal("media item must document copyable link formats")
	}
	permanentDelete := paths["/media/{id}/permanent"].(map[string]any)["delete"].(map[string]any)
	if permanentDelete["operationId"] != "permanentlyDeleteOwnMedia" {
		t.Fatalf("unexpected permanent-delete operation id: %v", permanentDelete["operationId"])
	}
	folderMove := paths["/media/{id}/folder"].(map[string]any)["patch"].(map[string]any)
	if folderMove["operationId"] != "moveOwnMediaToFolder" {
		t.Fatalf("unexpected media folder operation id: %v", folderMove["operationId"])
	}
}

func TestSpecDocumentsMemberCollectionRoutes(t *testing.T) {
	spec := Spec()
	paths := spec["paths"].(map[string]any)
	for _, path := range []string{"/me/folders", "/me/folders/{id}", "/me/albums", "/me/albums/{id}"} {
		if _, ok := paths[path]; !ok {
			t.Fatalf("missing member collection path %s", path)
		}
	}
	folders := paths["/me/folders"].(map[string]any)
	if folders["get"].(map[string]any)["operationId"] != "listOwnFolders" {
		t.Fatal("folders list must be owner-scoped")
	}
	albums := paths["/me/albums"].(map[string]any)
	if albums["post"].(map[string]any)["operationId"] != "createOwnAlbum" {
		t.Fatal("album create must be owner-scoped")
	}
	albumMedia, ok := paths["/me/albums/{id}/media"].(map[string]any)
	if !ok {
		t.Fatal("missing member album media path")
	}
	remove := albumMedia["delete"].(map[string]any)
	removeResponse := remove["responses"].(map[string]any)["200"].(map[string]any)
	removeContent := removeResponse["content"].(map[string]any)["application/json"].(map[string]any)
	if removeContent["x-schema-name"] != "AlbumMediaRemoval" {
		t.Fatalf("batch removal must use removal schema: %v", removeContent["x-schema-name"])
	}
	schemas := spec["components"].(map[string]any)["schemas"].(map[string]any)
	if _, ok := schemas["AlbumMediaRemoval"]; !ok {
		t.Fatal("missing album media removal schema")
	}
	order, ok := paths["/me/albums/{id}/media/order"].(map[string]any)
	if !ok || order["patch"].(map[string]any)["operationId"] != "reorderOwnAlbumMedia" {
		t.Fatal("album media order must be an owner-scoped patch contract")
	}
	move, ok := paths["/me/albums/{id}/media/move"].(map[string]any)
	if !ok || move["post"].(map[string]any)["operationId"] != "moveOwnAlbumMedia" {
		t.Fatal("album media move must be an owner-scoped post contract")
	}
	if _, ok := schemas["AlbumMediaMove"]; !ok {
		t.Fatal("missing album media move schema")
	}
}

func TestSpecDocumentsShareLinkBoundaries(t *testing.T) {
	paths := Spec()["paths"].(map[string]any)
	create := paths["/media/{id}/share-links"].(map[string]any)["post"].(map[string]any)
	if create["operationId"] != "createOwnShareLink" {
		t.Fatalf("unexpected share create operation id: %v", create["operationId"])
	}
	public := paths["/s/{token}"].(map[string]any)["get"].(map[string]any)
	if public["operationId"] != "resolvePublicShareLink" {
		t.Fatalf("unexpected public share operation id: %v", public["operationId"])
	}
	if _, ok := public["security"]; !ok {
		t.Fatal("public share endpoint must explicitly disable bearer security")
	}
	if _, ok := paths["/share-links/{id}"]; !ok {
		t.Fatal("missing share revoke route")
	}
}

func TestSpecDocumentsSharePasswordContract(t *testing.T) {
	paths := Spec()["paths"].(map[string]any)
	create := paths["/media/{id}/share-links"].(map[string]any)["post"].(map[string]any)
	body := create["requestBody"].(map[string]any)["content"].(map[string]any)["application/json"].(map[string]any)["schema"].(map[string]any)
	properties := body["properties"].(map[string]any)
	if _, ok := properties["password"]; !ok {
		t.Fatal("share creation must document an optional password")
	}
	public := paths["/s/{token}"].(map[string]any)["get"].(map[string]any)
	parameters := public["parameters"].([]map[string]any)
	for _, parameter := range parameters {
		if parameter["name"] == "password" && parameter["in"] == "query" {
			return
		}
	}
	t.Fatal("public share resolution must document the password query parameter")
}

func TestSpecDocumentsSignedURLAndHotlinkContracts(t *testing.T) {
	paths := Spec()["paths"].(map[string]any)
	for _, path := range []string{"/media/{id}/signed-url", "/media/{id}/hotlink-policy", "/hotlink-domains", "/hotlink-domains/{id}", "/i/{id}"} {
		if _, ok := paths[path]; !ok {
			t.Fatalf("missing link security path %s", path)
		}
	}
	signed := paths["/media/{id}/signed-url"].(map[string]any)["post"].(map[string]any)
	if signed["operationId"] != "createOwnSignedMediaURL" {
		t.Fatalf("unexpected signed URL operation id: %v", signed["operationId"])
	}
	policy := paths["/media/{id}/hotlink-policy"].(map[string]any)
	if _, ok := policy["get"]; !ok {
		t.Fatal("hotlink policy must be readable by the owner")
	}
	if _, ok := policy["put"]; !ok {
		t.Fatal("hotlink policy must be writable by the owner")
	}
	public := paths["/i/{id}"].(map[string]any)["get"].(map[string]any)
	if _, ok := public["security"]; !ok {
		t.Fatal("signed media delivery must explicitly disable bearer security")
	}
	schemas := Spec()["components"].(map[string]any)["schemas"].(map[string]any)
	for _, schema := range []string{"SignedMediaURL", "HotlinkPolicy", "HotlinkDomain"} {
		if _, ok := schemas[schema]; !ok {
			t.Fatalf("missing link security schema %s", schema)
		}
	}
}

func TestSpecDocumentsAdminMediaAccessLogContract(t *testing.T) {
	paths := Spec()["paths"].(map[string]any)
	path, ok := paths["/admin/media-access-logs"].(map[string]any)
	if !ok {
		t.Fatal("missing admin media access log path")
	}
	get, ok := path["get"].(map[string]any)
	if !ok || get["operationId"] != "listAdminMediaAccessLogs" {
		t.Fatalf("unexpected admin media access log operation: %v", get)
	}
	parameters := get["parameters"].([]map[string]any)
	for _, parameter := range parameters {
		if parameter["name"] == "signature" {
			t.Fatal("media access log contract must not expose signatures")
		}
	}
	schemas := Spec()["components"].(map[string]any)["schemas"].(map[string]any)
	accessLog, ok := schemas["MediaAccessLog"].(map[string]any)
	if !ok {
		t.Fatal("missing MediaAccessLog schema")
	}
	properties := accessLog["properties"].(map[string]any)
	for _, property := range []string{"id", "media_asset_id", "variant", "delivery_mode", "result", "referer_host", "accessed_at"} {
		if _, ok := properties[property]; !ok {
			t.Fatalf("MediaAccessLog must expose %s", property)
		}
	}
	for _, sensitive := range []string{"signature", "token", "password"} {
		if _, ok := properties[sensitive]; ok {
			t.Fatalf("MediaAccessLog must not expose %s", sensitive)
		}
	}
}

func TestSpecDocumentsPersonalAPITokenBoundaries(t *testing.T) {
	paths := Spec()["paths"].(map[string]any)
	for _, path := range []string{"/tokens", "/tokens/{id}", "/tokens/{id}/rotate"} {
		if _, ok := paths[path]; !ok {
			t.Fatalf("missing Personal API Token path %s", path)
		}
	}
	create := paths["/tokens"].(map[string]any)["post"].(map[string]any)
	if create["operationId"] != "createPersonalAPIToken" {
		t.Fatalf("unexpected token create operation id: %v", create["operationId"])
	}
	schemas := Spec()["components"].(map[string]any)["schemas"].(map[string]any)
	if _, ok := schemas["PersonalAPIToken"]; !ok {
		t.Fatal("missing PersonalAPIToken schema")
	}
	request := create["requestBody"].(map[string]any)["content"].(map[string]any)["application/json"].(map[string]any)["schema"].(map[string]any)
	if _, ok := request["properties"].(map[string]any)["expires_at"]; !ok {
		t.Fatal("token create request must document expiry")
	}
}
