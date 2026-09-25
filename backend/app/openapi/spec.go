package openapi

func Spec() map[string]any {
	return map[string]any{
		"openapi":  "3.0.3",
		"info":     map[string]any{"title": "Go Vue Admin API", "version": "0.1.0", "description": "The checked-in API contract for the admin backend."},
		"servers":  []map[string]any{{"url": "/api/v1"}},
		"security": []map[string]any{{"bearerAuth": []any{}}},
		"components": map[string]any{
			"securitySchemes": map[string]any{"bearerAuth": map[string]any{"type": "http", "scheme": "bearer", "bearerFormat": "Personal API Token or session"}, "apiKeyAuth": map[string]any{"type": "apiKey", "in": "header", "name": "X-API-Key"}},
			"schemas": map[string]any{
				"UserStatus":                      map[string]any{"type": "string", "enum": []string{"active", "disabled", "locked"}},
				"DataScope":                       map[string]any{"type": "string", "enum": []string{"all", "own"}},
				"FieldPermissionOverride":         map[string]any{"type": "object", "required": []string{"readable", "writable"}, "properties": map[string]any{"readable": map[string]any{"type": "boolean"}, "writable": map[string]any{"type": "boolean"}}},
				"Error":                           map[string]any{"type": "object", "required": []string{"code", "request_id", "retryable"}, "properties": map[string]any{"code": map[string]any{"type": "string"}, "request_id": map[string]any{"type": "string"}, "retryable": map[string]any{"type": "boolean"}}},
				"ResourceOption":                  resourceOptionSchema(),
				"ResourceField":                   resourceFieldSchema(),
				"ResourceFilter":                  resourceFilterSchema(),
				"ResourceColumn":                  resourceColumnSchema(),
				"ResourceAction":                  resourceActionSchema(),
				"ActionPayloadField":              actionPayloadFieldSchema(),
				"ResourceRelation":                resourceRelationSchema(),
				"ResourceFormGroup":               resourceFormGroupSchema(),
				"ResourceDetailSection":           resourceDetailSectionSchema(),
				"ResourceFieldDependency":         resourceFieldDependencySchema(),
				"RelationOption":                  relationOptionSchema(),
				"RelationOptionList":              relationOptionListSchema(),
				"ResourceListMeta":                resourceListMetaSchema(),
				"ResourceManifest":                resourceManifestSchema(),
				"ActionRequest":                   actionRequestSchema(),
				"ActionResponse":                  actionResponseSchema(),
				"GlobalSearchResult":              globalSearchResultSchema(),
				"GlobalSearchResponse":            globalSearchResponseSchema(),
				"AuditCleanupMode":                map[string]any{"type": "string", "enum": []string{"retention", "selected", "filtered", "all"}},
				"AuditCleanupRequest":             auditCleanupRequestSchema(),
				"AuditCleanupResponse":            auditCleanupResponseSchema(),
				"Notification":                    notificationSchema(),
				"NotificationReadResponse":        notificationReadResponseSchema(),
				"NotificationMarkAllReadResponse": notificationMarkAllReadResponseSchema(),
				"NotificationUnreadCount":         notificationUnreadCountSchema(),
				"Plan":                            planSchema(),
				"Subscription":                    subscriptionSchema(),
				"MediaVariant":                    mediaVariantSchema(),
				"MediaItem":                       mediaItemSchema(),
				"MediaListResponse":               mediaListResponseSchema(),
				"MediaTrashConfirmation":          map[string]any{"type": "object", "required": []string{"confirm"}, "properties": map[string]any{"confirm": map[string]any{"type": "string", "enum": []string{"empty-trash"}}}},
				"MediaTrashEmptyResponse":         map[string]any{"type": "object", "required": []string{"deleted_count"}, "properties": map[string]any{"deleted_count": map[string]any{"type": "integer", "format": "int64", "minimum": 0}}},
				"MediaUploadResponse":             mediaUploadResponseSchema(),
				"BatchMediaUploadResponse":        batchMediaUploadResponseSchema(),
				"Collection":                      collectionSchema(),
				"AlbumMediaMutation":              map[string]any{"type": "object", "required": []string{"added_ids", "skipped_ids"}, "properties": map[string]any{"added_ids": map[string]any{"type": "array", "items": map[string]any{"type": "integer", "format": "int64"}}, "skipped_ids": map[string]any{"type": "array", "items": map[string]any{"type": "integer", "format": "int64"}}}},
				"AlbumMediaRemoval":               map[string]any{"type": "object", "required": []string{"removed_ids", "skipped_ids"}, "properties": map[string]any{"removed_ids": map[string]any{"type": "array", "items": map[string]any{"type": "integer", "format": "int64"}}, "skipped_ids": map[string]any{"type": "array", "items": map[string]any{"type": "integer", "format": "int64"}}}},
				"AlbumMediaMove":                  map[string]any{"type": "object", "required": []string{"moved_ids", "skipped_ids"}, "properties": map[string]any{"moved_ids": map[string]any{"type": "array", "items": map[string]any{"type": "integer", "format": "int64"}}, "skipped_ids": map[string]any{"type": "array", "items": map[string]any{"type": "integer", "format": "int64"}}}},
				"AlbumMediaIDsResponse":           map[string]any{"type": "object", "required": []string{"data", "meta"}, "properties": map[string]any{"data": map[string]any{"type": "array", "items": map[string]any{"type": "integer", "format": "int64"}}, "meta": map[string]any{"type": "object", "properties": map[string]any{"album_id": map[string]any{"type": "integer", "format": "int64"}, "total": map[string]any{"type": "integer"}}}}},
				"ShareLink":                       shareLinkSchema(),
				"SignedMediaURL":                  signedMediaURLSchema(),
				"HotlinkPolicy":                   hotlinkPolicySchema(),
				"HotlinkDomain":                   hotlinkDomainSchema(),
				"MediaAccessLog":                  mediaAccessLogSchema(),
				"PersonalAPIToken":                personalAPITokenSchema(),
				"AdSlot":                          adSlotSchema(),
				"AdSlotListResponse":              adSlotListResponseSchema(),
				"DiscoveryStatus":                 discoveryStatusSchema(),
				"DiscoveryFeedItem":               discoveryFeedItemSchema(),
				"DiscoveryFeedResponse":           discoveryFeedResponseSchema(),
				"DiscoverySubmissionResponse":     discoverySubmissionResponseSchema(),
				"SitePresentation":                map[string]any{"type": "object", "required": []string{"watermark_fallback_image_url"}, "properties": map[string]any{"watermark_fallback_image_url": map[string]any{"type": "string", "format": "uri", "description": "Safe public fallback image URL; empty when the built-in member fallback is used."}}},
			},
		},
		"paths": map[string]any{
			"/auth/login": map[string]any{"post": map[string]any{
				"security": []any{}, "operationId": "login", "requestBody": jsonBody("LoginRequest", map[string]any{"type": "object", "required": []string{"email", "password"}, "properties": map[string]any{"email": map[string]any{"type": "string", "format": "email"}, "password": map[string]any{"type": "string", "format": "password"}}}),
				"responses": map[string]any{"200": jsonResponse("LoginResponse"), "403": errorResponse(), "429": errorResponse(), "503": errorResponse()},
			}},
			"/auth/refresh": map[string]any{"post": map[string]any{
				"security": []any{}, "operationId": "refresh", "responses": map[string]any{"200": jsonResponse("RefreshResponse"), "401": errorResponse()},
			}},
			"/auth/logout-all":              map[string]any{"post": operation("logoutAll")},
			"/auth/me":                      map[string]any{"get": operation("currentUser")},
			"/plans":                        map[string]any{"get": operation("listPlans")},
			"/site/presentation":            map[string]any{"get": map[string]any{"security": []any{}, "operationId": "getSitePresentation", "responses": map[string]any{"200": jsonResponse("SitePresentation")}}},
			"/discovery/status":             map[string]any{"get": map[string]any{"security": []any{}, "operationId": "getDiscoveryStatus", "responses": map[string]any{"200": jsonResponse("DiscoveryStatus"), "500": errorResponse()}}},
			"/discovery/feed":               map[string]any{"get": map[string]any{"security": []any{}, "operationId": "listDiscoveryFeed", "parameters": []map[string]any{queryParameter("page", "integer"), queryParameter("per_page", "integer")}, "responses": map[string]any{"200": jsonResponse("DiscoveryFeedResponse"), "404": errorResponse(), "500": errorResponse()}}},
			"/discovery/media/{id}/content": map[string]any{"get": map[string]any{"security": []any{}, "operationId": "getDiscoveryMediaContent", "parameters": []map[string]any{pathParameter("id"), {"name": "variant", "in": "query", "schema": map[string]any{"type": "string", "enum": []string{"original", "thumbnail", "medium"}, "default": "thumbnail"}}}, "responses": map[string]any{"200": map[string]any{"description": "Public discovery media bytes", "content": map[string]any{"image/*": map[string]any{"schema": map[string]any{"type": "string", "format": "binary"}}}}, "404": errorResponse()}}},
			"/payment-gateways":             map[string]any{"get": map[string]any{"operationId": "listAvailablePaymentGateways", "responses": map[string]any{"200": map[string]any{"description": "Available configured gateway codes"}, "401": errorResponse()}}},
			"/ads": map[string]any{"get": map[string]any{
				"operationId": "listPublishedAds",
				"parameters":  []map[string]any{{"name": "placement", "in": "query", "required": false, "schema": map[string]any{"type": "string", "enum": []string{"header", "footer", "left", "right"}}}},
				"responses":   map[string]any{"200": jsonResponse("AdSlotListResponse"), "401": errorResponse(), "422": errorResponse(), "500": errorResponse()},
			}},
			"/subscription":    map[string]any{"get": operation("currentSubscription")},
			"/me/usage":        map[string]any{"get": operation("currentUsage")},
			"/me/usage/ledger": map[string]any{"get": listOperation("listOwnUsageLedger", []map[string]any{queryParameter("page", "integer"), queryParameter("per_page", "integer")})},
			"/me/folders": map[string]any{
				"get": collectionListOperation("listOwnFolders"), "post": collectionWriteOperation("createOwnFolder", "201"),
			},
			"/me/folders/{id}": map[string]any{
				"patch": collectionWriteOperation("updateOwnFolder", "200", true), "delete": collectionDeleteOperation("deleteOwnFolder"),
			},
			"/me/albums": map[string]any{
				"get": collectionListOperation("listOwnAlbums"), "post": collectionWriteOperation("createOwnAlbum", "201"),
			},
			"/me/albums/{id}": map[string]any{
				"patch": collectionWriteOperation("updateOwnAlbum", "200", true), "delete": collectionDeleteOperation("deleteOwnAlbum"),
			},
			"/me/albums/{id}/media": map[string]any{
				"get":    map[string]any{"operationId": "listOwnAlbumMedia", "parameters": []map[string]any{pathParameter("id")}, "responses": map[string]any{"200": jsonResponse("AlbumMediaIDsResponse"), "401": errorResponse(), "404": errorResponse()}},
				"post":   map[string]any{"operationId": "addOwnMediaToAlbum", "parameters": []map[string]any{pathParameter("id")}, "requestBody": jsonBody("AlbumMediaRequest", map[string]any{"type": "object", "required": []string{"media_ids"}, "properties": map[string]any{"media_ids": map[string]any{"type": "array", "minItems": 1, "maxItems": 100, "items": map[string]any{"type": "integer", "format": "int64", "minimum": 1}}}}), "responses": map[string]any{"201": jsonResponse("AlbumMediaMutation"), "401": errorResponse(), "404": errorResponse(), "422": errorResponse()}},
				"delete": map[string]any{"operationId": "removeOwnMediaBatchFromAlbum", "parameters": []map[string]any{pathParameter("id")}, "requestBody": jsonBody("AlbumMediaRequest", map[string]any{"type": "object", "required": []string{"media_ids"}, "properties": map[string]any{"media_ids": map[string]any{"type": "array", "minItems": 1, "maxItems": 100, "items": map[string]any{"type": "integer", "format": "int64", "minimum": 1}}}}), "responses": map[string]any{"200": jsonResponse("AlbumMediaRemoval"), "401": errorResponse(), "404": errorResponse(), "422": errorResponse()}},
			},
			"/me/albums/{id}/media/order": map[string]any{
				"patch": map[string]any{"operationId": "reorderOwnAlbumMedia", "parameters": []map[string]any{pathParameter("id")}, "requestBody": jsonBody("AlbumMediaRequest", map[string]any{"type": "object", "required": []string{"media_ids"}, "properties": map[string]any{"media_ids": map[string]any{"type": "array", "minItems": 1, "maxItems": 100, "items": map[string]any{"type": "integer", "format": "int64", "minimum": 1}}}}), "responses": map[string]any{"200": jsonResponse("AlbumMediaIDsResponse"), "401": errorResponse(), "404": errorResponse(), "422": errorResponse()}},
			},
			"/me/albums/{id}/media/move": map[string]any{
				"post": map[string]any{"operationId": "moveOwnAlbumMedia", "parameters": []map[string]any{pathParameter("id")}, "requestBody": jsonBody("AlbumMediaMoveRequest", map[string]any{"type": "object", "required": []string{"source_album_id", "media_ids"}, "properties": map[string]any{"source_album_id": map[string]any{"type": "integer", "format": "int64", "minimum": 1}, "media_ids": map[string]any{"type": "array", "minItems": 1, "maxItems": 100, "items": map[string]any{"type": "integer", "format": "int64", "minimum": 1}}}}), "responses": map[string]any{"200": jsonResponse("AlbumMediaMove"), "401": errorResponse(), "404": errorResponse(), "422": errorResponse()}},
			},
			"/me/albums/{id}/media/{media_id}": map[string]any{"delete": map[string]any{"operationId": "removeOwnMediaFromAlbum", "parameters": []map[string]any{pathParameter("id"), pathParameter("media_id")}, "responses": map[string]any{"204": map[string]any{"description": "Media removed from album"}, "401": errorResponse(), "404": errorResponse()}}},
			"/uploads": map[string]any{"post": map[string]any{
				"operationId": "createMediaUpload",
				"parameters":  []map[string]any{{"name": "Idempotency-Key", "in": "header", "required": true, "schema": map[string]any{"type": "string", "minLength": 8, "maxLength": 160}}},
				"requestBody": multipartUploadBody(),
				"responses":   map[string]any{"201": jsonResponse("MediaUploadResponse"), "202": jsonResponse("MediaUploadResponse"), "400": errorResponse(), "401": errorResponse(), "409": errorResponse(), "413": errorResponse(), "422": errorResponse(), "429": errorResponse(), "507": errorResponse()},
			}},
			"/upload": map[string]any{"post": map[string]any{
				"operationId": "createCompatibilityMediaUpload", "requestBody": multipartUploadBody(),
				"responses": map[string]any{"201": jsonResponse("MediaUploadResponse"), "202": jsonResponse("MediaUploadResponse"), "401": errorResponse(), "403": errorResponse(), "413": errorResponse(), "422": errorResponse()},
			}},
			"/images": map[string]any{"get": listOperation("listCompatibilityOwnMedia", []map[string]any{queryParameter("page", "integer"), queryParameter("per_page", "integer")})},
			"/image/{id}": map[string]any{
				"get":    map[string]any{"operationId": "getCompatibilityOwnMedia", "parameters": []map[string]any{pathParameter("id")}, "responses": map[string]any{"200": jsonResponse("MediaItem"), "401": errorResponse(), "403": errorResponse(), "404": errorResponse()}},
				"delete": map[string]any{"operationId": "deleteCompatibilityOwnMedia", "parameters": []map[string]any{pathParameter("id")}, "responses": map[string]any{"200": jsonResponse("MediaItem"), "401": errorResponse(), "403": errorResponse(), "404": errorResponse()}},
			},
			"/uploads/batch": map[string]any{"post": map[string]any{
				"operationId": "createBatchMediaUpload",
				"parameters":  []map[string]any{{"name": "Idempotency-Key", "in": "header", "required": true, "schema": map[string]any{"type": "string", "minLength": 8, "maxLength": 160}}},
				"requestBody": batchMultipartUploadBody(),
				"responses":   map[string]any{"207": jsonResponse("BatchMediaUploadResponse"), "400": errorResponse(), "401": errorResponse(), "413": errorResponse(), "422": errorResponse(), "429": errorResponse(), "507": errorResponse()},
			}},
			"/uploads/{id}": map[string]any{"get": map[string]any{
				"operationId": "getOwnUploadStatus", "parameters": []map[string]any{pathParameter("id")},
				"responses": map[string]any{"200": jsonResponse("MediaUploadResponse"), "202": jsonResponse("MediaUploadResponse"), "401": errorResponse(), "404": errorResponse()},
			}},
			"/uploads/{id}/retry": map[string]any{"post": map[string]any{
				"operationId": "retryOwnUploadSession", "parameters": []map[string]any{pathParameter("id")},
				"responses": map[string]any{"200": jsonResponse("MediaUploadResponse"), "202": jsonResponse("MediaUploadResponse"), "401": errorResponse(), "404": errorResponse(), "409": errorResponse(), "503": errorResponse()},
			}},
			"/media": map[string]any{"get": listOperation("listOwnMedia", []map[string]any{
				queryParameter("page", "integer"), queryParameter("per_page", "integer"),
				{"name": "state", "in": "query", "schema": map[string]any{"type": "string", "enum": []string{"ready", "trash"}}},
				queryParameter("search", "string"),
			})},
			"/media/{id}/content": map[string]any{"get": map[string]any{
				"operationId": "getOwnMediaContent",
				"parameters":  []map[string]any{pathParameter("id"), {"name": "variant", "in": "query", "schema": map[string]any{"type": "string", "enum": []string{"original", "thumbnail", "medium"}}}},
				"responses":   map[string]any{"200": map[string]any{"description": "Private media bytes", "content": map[string]any{"image/jpeg": map[string]any{"schema": map[string]any{"type": "string", "format": "binary"}}, "image/png": map[string]any{"schema": map[string]any{"type": "string", "format": "binary"}}, "image/gif": map[string]any{"schema": map[string]any{"type": "string", "format": "binary"}}}}, "401": errorResponse(), "404": errorResponse()},
			}},
			"/media/trash": map[string]any{"delete": map[string]any{
				"operationId": "emptyOwnMediaTrash",
				"requestBody": jsonBody("MediaTrashConfirmation", map[string]any{"type": "object", "required": []string{"confirm"}, "properties": map[string]any{
					"confirm": map[string]any{"type": "string", "enum": []string{"empty-trash"}},
				}}),
				"responses": map[string]any{"200": jsonResponse("MediaTrashEmptyResponse"), "401": errorResponse(), "422": errorResponse()},
			}},
			"/media/{id}": map[string]any{
				"get": map[string]any{"operationId": "getOwnMediaDetails", "parameters": []map[string]any{pathParameter("id")}, "responses": map[string]any{"200": jsonResponse("MediaItem"), "401": errorResponse(), "404": errorResponse()}},
				"delete": map[string]any{
					"operationId": "moveOwnMediaToTrash", "parameters": []map[string]any{pathParameter("id")},
					"responses": map[string]any{"200": jsonResponse("MediaItem"), "401": errorResponse(), "404": errorResponse()},
				},
			},
			"/media/{id}/permanent": map[string]any{"delete": map[string]any{
				"operationId": "permanentlyDeleteOwnMedia", "parameters": []map[string]any{pathParameter("id")},
				"requestBody": jsonBody("PermanentMediaDeleteRequest", map[string]any{"type": "object", "required": []string{"confirm"}, "properties": map[string]any{
					"confirm": map[string]any{"type": "string", "enum": []string{"permanently-delete"}},
				}}),
				"responses": map[string]any{"200": jsonResponse("MediaItem"), "401": errorResponse(), "404": errorResponse(), "409": errorResponse(), "422": errorResponse(), "500": errorResponse()},
			}},
			"/media/{id}/restore": map[string]any{"post": map[string]any{
				"operationId": "restoreOwnMedia", "parameters": []map[string]any{pathParameter("id")},
				"responses": map[string]any{"200": jsonResponse("MediaItem"), "401": errorResponse(), "404": errorResponse()},
			}},
			"/media/{id}/folder": map[string]any{"patch": map[string]any{
				"operationId": "moveOwnMediaToFolder", "parameters": []map[string]any{pathParameter("id")},
				"requestBody": jsonBody("MediaFolderRequest", map[string]any{"type": "object", "properties": map[string]any{"folder_id": map[string]any{"type": "integer", "format": "int64", "nullable": true}}}),
				"responses":   map[string]any{"200": jsonResponse("MediaFolderResponse"), "401": errorResponse(), "404": errorResponse(), "422": errorResponse()},
			}},
			"/media/{id}/discovery-submit": map[string]any{"post": map[string]any{"operationId": "submitOwnMediaToDiscovery", "parameters": []map[string]any{pathParameter("id")}, "responses": map[string]any{"202": jsonResponse("DiscoverySubmissionResponse"), "401": errorResponse(), "404": errorResponse(), "409": errorResponse()}}},
			"/media/{id}/share-links": map[string]any{"post": map[string]any{
				"operationId": "createOwnShareLink", "parameters": []map[string]any{pathParameter("id")},
				"requestBody": jsonBody("ShareLinkRequest", map[string]any{"type": "object", "properties": map[string]any{"expires_at": map[string]any{"type": "string", "format": "date-time", "nullable": true}, "password": map[string]any{"type": "string", "format": "password", "minLength": 8, "maxLength": 72, "description": "Optional password; the server stores only a hash."}}}),
				"responses":   map[string]any{"201": jsonResponse("ShareLink"), "401": errorResponse(), "404": errorResponse(), "422": errorResponse()},
			}},
			"/media/{id}/signed-url": map[string]any{"post": map[string]any{
				"operationId": "createOwnSignedMediaURL", "parameters": []map[string]any{pathParameter("id")},
				"requestBody": jsonBody("SignedMediaURLRequest", map[string]any{"type": "object", "properties": map[string]any{"variant": map[string]any{"type": "string", "enum": []string{"original", "thumbnail", "medium"}}, "expires_in": map[string]any{"type": "integer", "minimum": 60, "maximum": 86400, "default": 600}}}),
				"responses":   map[string]any{"201": jsonResponse("SignedMediaURL"), "401": errorResponse(), "404": errorResponse(), "422": errorResponse(), "503": errorResponse()},
			}},
			"/media/{id}/hotlink-policy": map[string]any{
				"get": map[string]any{"operationId": "getOwnHotlinkPolicy", "parameters": []map[string]any{pathParameter("id")}, "responses": map[string]any{"200": jsonResponse("HotlinkPolicy"), "401": errorResponse(), "404": errorResponse()}},
				"put": map[string]any{"operationId": "updateOwnHotlinkPolicy", "parameters": []map[string]any{pathParameter("id")}, "requestBody": jsonBody("HotlinkPolicyRequest", map[string]any{"type": "object", "required": []string{"mode"}, "properties": map[string]any{"mode": map[string]any{"type": "string", "enum": []string{"off", "referer", "signed", "hybrid"}}, "allow_no_referer": map[string]any{"type": "boolean", "default": false}}}), "responses": map[string]any{"200": jsonResponse("HotlinkPolicy"), "401": errorResponse(), "404": errorResponse(), "422": errorResponse()}},
			},
			"/hotlink-domains": map[string]any{
				"get":  map[string]any{"operationId": "listOwnHotlinkDomains", "responses": map[string]any{"200": jsonResponse("HotlinkDomainList"), "401": errorResponse()}},
				"post": map[string]any{"operationId": "createOwnHotlinkDomain", "requestBody": jsonBody("HotlinkDomainRequest", map[string]any{"type": "object", "required": []string{"domain"}, "properties": map[string]any{"domain": map[string]any{"type": "string", "maxLength": 255}}}), "responses": map[string]any{"201": jsonResponse("HotlinkDomain"), "401": errorResponse(), "409": errorResponse(), "422": errorResponse()}},
			},
			"/hotlink-domains/{id}": map[string]any{"delete": map[string]any{"operationId": "deleteOwnHotlinkDomain", "parameters": []map[string]any{pathParameter("id")}, "responses": map[string]any{"204": map[string]any{"description": "Domain revoked"}, "401": errorResponse(), "404": errorResponse()}}},
			"/share-links": map[string]any{"get": map[string]any{
				"operationId": "listOwnShareLinks", "responses": map[string]any{"200": map[string]any{"description": "Own share links", "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"type": "object", "required": []string{"data"}, "properties": map[string]any{"data": map[string]any{"type": "array", "items": map[string]any{"$ref": "#/components/schemas/ShareLink"}}}}}}}, "401": errorResponse()},
			}},
			"/share-links/{id}": map[string]any{"delete": map[string]any{
				"operationId": "revokeOwnShareLink", "parameters": []map[string]any{pathParameter("id")},
				"responses": map[string]any{"204": map[string]any{"description": "Share link revoked"}, "401": errorResponse(), "404": errorResponse()},
			}},
			"/tokens": map[string]any{
				"get": listOperation("listPersonalAPITokens", nil),
				"post": map[string]any{
					"operationId": "createPersonalAPIToken",
					"requestBody": jsonBody("PersonalAPITokenRequest", map[string]any{"type": "object", "required": []string{"name"}, "properties": map[string]any{
						"name":       map[string]any{"type": "string", "maxLength": 120},
						"scopes":     map[string]any{"type": "array", "deprecated": true, "description": "Ignored for new tokens; the server assigns upload:write, media:read and media:delete.", "items": map[string]any{"type": "string"}},
						"expires_at": map[string]any{"type": "string", "format": "date-time", "nullable": true},
					}}),
					"responses": map[string]any{"201": jsonResponse("PersonalAPITokenCreated"), "401": errorResponse(), "422": errorResponse()},
				},
			},
			"/tokens/{id}": map[string]any{"delete": map[string]any{
				"operationId": "revokePersonalAPIToken", "parameters": []map[string]any{pathParameter("id")},
				"responses": map[string]any{"204": map[string]any{"description": "Token revoked"}, "401": errorResponse(), "404": errorResponse()},
			}},
			"/tokens/{id}/rotate": map[string]any{"post": map[string]any{
				"operationId": "rotatePersonalAPIToken", "parameters": []map[string]any{pathParameter("id")},
				"responses": map[string]any{"201": jsonResponse("PersonalAPITokenCreated"), "401": errorResponse(), "404": errorResponse()},
			}},
			"/s/{token}": map[string]any{"get": map[string]any{
				"security": []any{}, "operationId": "resolvePublicShareLink", "parameters": []map[string]any{pathParameter("token"), {"name": "variant", "in": "query", "schema": map[string]any{"type": "string", "enum": []string{"original", "thumbnail", "medium"}}}, {"name": "password", "in": "query", "schema": map[string]any{"type": "string", "format": "password"}, "description": "Required only for password-protected share links."}},
				"responses": map[string]any{"200": map[string]any{"description": "Shared media bytes", "content": map[string]any{"image/jpeg": map[string]any{"schema": map[string]any{"type": "string", "format": "binary"}}, "image/png": map[string]any{"schema": map[string]any{"type": "string", "format": "binary"}}, "image/gif": map[string]any{"schema": map[string]any{"type": "string", "format": "binary"}}}}, "401": errorResponse(), "404": errorResponse()},
			}},
			"/i/{id}": map[string]any{"get": map[string]any{
				"security": []any{}, "operationId": "resolveSignedMediaURL", "parameters": []map[string]any{pathParameter("id"), {"name": "variant", "in": "query", "required": true, "schema": map[string]any{"type": "string", "enum": []string{"original", "thumbnail", "medium"}}}, {"name": "expires", "in": "query", "required": true, "schema": map[string]any{"type": "integer", "format": "int64"}}, {"name": "signature", "in": "query", "required": true, "schema": map[string]any{"type": "string"}}},
				"responses": map[string]any{"200": map[string]any{"description": "Signed media bytes", "content": map[string]any{"image/*": map[string]any{"schema": map[string]any{"type": "string", "format": "binary"}}}}, "404": errorResponse()},
			}},
			"/admin/registry": map[string]any{"get": operation("listResources")},
			"/admin/search":   map[string]any{"get": globalSearchOperation()},
			"/admin/{resource}": map[string]any{
				"get":  listOperation("listResourceRows", []map[string]any{pathParameter("resource"), queryParameter("page", "integer"), queryParameter("per_page", "integer"), queryParameter("search", "string"), queryParameter("sort", "string"), queryParameter("dir", "string"), queryParameter("trashed", "string")}),
				"post": resourceWriteOperation("createResource", "201"),
			},
			"/admin/{resource}/export":                       exportOperation(),
			"/admin/{resource}/actions/{action}":             map[string]any{"post": resourceActionOperation()},
			"/admin/{resource}/relations/{relation}/options": map[string]any{"get": relationOptionsOperation()},
			"/admin/{resource}/{id}/relations/{relation}":    map[string]any{"get": relationRecordsOperation()},
			"/admin/{resource}/{id}": map[string]any{
				"get":    resourceItemOperation("showResource"),
				"put":    resourceWriteOperation("updateResource", "200", true),
				"delete": map[string]any{"operationId": "deleteResource", "parameters": []map[string]any{pathParameter("resource"), pathParameter("id")}, "responses": map[string]any{"204": map[string]any{"description": "Resource deleted"}, "401": errorResponse(), "403": errorResponse(), "404": errorResponse()}},
			},
			"/admin/overview":             map[string]any{"get": operation("adminOverview")},
			"/admin/audit-logs":           map[string]any{"get": listOperation("auditLogs", []map[string]any{queryParameter("page", "integer"), queryParameter("per_page", "integer"), queryParameter("action", "string"), queryParameter("user_id", "integer")})},
			"/admin/media-access-logs":    map[string]any{"get": listOperation("listAdminMediaAccessLogs", []map[string]any{queryParameter("page", "integer"), queryParameter("per_page", "integer"), queryParameter("media_id", "integer"), queryParameter("variant", "string"), queryParameter("delivery_mode", "string"), queryParameter("result", "string"), queryParameter("referer_host", "string")})},
			"/admin/audit-logs/cleanup":   map[string]any{"post": map[string]any{"operationId": "cleanupAuditLogs", "requestBody": jsonBody("AuditCleanupRequest", map[string]any{"$ref": "#/components/schemas/AuditCleanupRequest"}), "responses": map[string]any{"200": jsonResponse("AuditCleanupResponse"), "401": errorResponse(), "403": errorResponse(), "422": errorResponse()}}},
			"/notifications":              map[string]any{"get": listOperation("notifications", []map[string]any{queryParameter("page", "integer"), queryParameter("per_page", "integer"), queryParameter("unread", "boolean")})},
			"/notifications/unread-count": map[string]any{"get": operation("notificationUnreadCount")},
			"/notifications/{id}/read":    map[string]any{"put": map[string]any{"operationId": "markNotificationRead", "parameters": []map[string]any{pathParameter("id")}, "responses": map[string]any{"200": jsonResponse("NotificationReadResponse"), "401": errorResponse(), "404": errorResponse(), "422": errorResponse()}}},
			"/notifications/read-all":     map[string]any{"put": map[string]any{"operationId": "markAllNotificationsRead", "responses": map[string]any{"200": jsonResponse("NotificationMarkAllReadResponse"), "401": errorResponse()}}},
			"/admin/settings":             map[string]any{"get": operation("systemSettings")},
			"/admin/settings/{key}":       map[string]any{"put": map[string]any{"operationId": "updateSystemSetting", "parameters": []map[string]any{{"name": "key", "in": "path", "required": true, "schema": map[string]any{"type": "string"}}}, "requestBody": jsonBody("SystemSettingRequest", map[string]any{"type": "object", "required": []string{"value"}, "properties": map[string]any{"value": map[string]any{"type": "string"}, "value_type": map[string]any{"type": "string", "enum": []string{"string", "boolean", "integer", "json"}}, "group": map[string]any{"type": "string"}, "description": map[string]any{"type": "string"}}}), "responses": map[string]any{"200": jsonResponse("SystemSettingResponse"), "422": errorResponse()}}},
			"/admin/roles/{id}/permissions": map[string]any{"put": map[string]any{
				"operationId": "replaceRolePermissions",
				"parameters":  []map[string]any{pathParameter("id")},
				"requestBody": jsonBody("RolePermissionsRequest", map[string]any{"type": "object", "required": []string{"permission_ids"}, "properties": map[string]any{
					"permission_ids": map[string]any{"type": "array", "items": map[string]any{"type": "integer", "format": "int64"}},
					"scopes":         map[string]any{"type": "object", "additionalProperties": map[string]any{"$ref": "#/components/schemas/DataScope"}},
					"fields":         map[string]any{"type": "object", "additionalProperties": map[string]any{"type": "object", "additionalProperties": map[string]any{"$ref": "#/components/schemas/FieldPermissionOverride"}}},
				}}),
				"responses": map[string]any{"204": map[string]any{"description": "Permissions replaced"}, "401": errorResponse(), "403": errorResponse(), "404": errorResponse(), "422": errorResponse()},
			}},
		},
	}
}

func planSchema() map[string]any {
	return map[string]any{"type": "object", "required": []string{"id", "code", "name", "price_amount", "currency", "billing_period", "status"}, "properties": map[string]any{
		"id": map[string]any{"type": "integer", "format": "int64"}, "code": map[string]any{"type": "string"}, "name": map[string]any{"type": "string"},
		"description": map[string]any{"type": "string", "nullable": true}, "price_amount": map[string]any{"type": "integer", "format": "int64"},
		"currency": map[string]any{"type": "string"}, "billing_period": map[string]any{"type": "string"}, "status": map[string]any{"type": "string"},
		"entitlements": map[string]any{"type": "object", "additionalProperties": true}, "sort_order": map[string]any{"type": "integer"},
	}}
}

func subscriptionSchema() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{
		"subscription": map[string]any{"type": "object", "properties": map[string]any{"id": map[string]any{"type": "integer"}, "user_id": map[string]any{"type": "integer"}, "plan_id": map[string]any{"type": "integer"}, "status": map[string]any{"type": "string"}}},
		"plan":         map[string]any{"$ref": "#/components/schemas/Plan"},
	}}
}

func mediaVariantSchema() map[string]any {
	return map[string]any{"type": "object", "required": []string{"url", "content_type", "size_bytes", "width", "height"}, "properties": map[string]any{
		"url": map[string]any{"type": "string"}, "content_type": map[string]any{"type": "string"},
		"size_bytes": map[string]any{"type": "integer", "format": "int64"},
		"width":      map[string]any{"type": "integer"}, "height": map[string]any{"type": "integer"},
	}}
}

func mediaItemSchema() map[string]any {
	return map[string]any{"type": "object", "required": []string{"id", "original_name", "content_type", "size_bytes", "width", "height", "status", "variants"}, "properties": map[string]any{
		"id": map[string]any{"type": "integer", "format": "int64"}, "original_name": map[string]any{"type": "string"},
		"folder_id":    map[string]any{"type": "integer", "format": "int64", "nullable": true},
		"content_type": map[string]any{"type": "string"}, "format": map[string]any{"type": "string"},
		"size_bytes": map[string]any{"type": "integer", "format": "int64"}, "width": map[string]any{"type": "integer"},
		"height": map[string]any{"type": "integer"}, "status": map[string]any{"type": "string", "enum": []string{"ready", "deleted", "cleanup_pending", "physically_deleted"}},
		"created_at": map[string]any{"type": "string", "format": "date-time"},
		"links":      map[string]any{"type": "object", "nullable": true, "additionalProperties": map[string]any{"type": "string"}},
		"variants":   map[string]any{"type": "object", "additionalProperties": map[string]any{"$ref": "#/components/schemas/MediaVariant"}},
	}}
}

func mediaListResponseSchema() map[string]any {
	return map[string]any{"type": "object", "required": []string{"data", "meta"}, "properties": map[string]any{
		"data": map[string]any{"type": "array", "items": map[string]any{"$ref": "#/components/schemas/MediaItem"}},
		"meta": map[string]any{"type": "object", "properties": map[string]any{"page": map[string]any{"type": "integer"}, "per_page": map[string]any{"type": "integer"}, "total": map[string]any{"type": "integer"}}},
	}}
}

func discoveryStatusSchema() map[string]any {
	return map[string]any{"type": "object", "required": []string{"enabled", "submissions_enabled"}, "properties": map[string]any{
		"enabled": map[string]any{"type": "boolean"}, "submissions_enabled": map[string]any{"type": "boolean"},
	}}
}

func discoveryFeedItemSchema() map[string]any {
	return map[string]any{"type": "object", "required": []string{"id", "original_name", "content_type", "width", "height", "thumbnail_url", "original_url"}, "properties": map[string]any{
		"id": map[string]any{"type": "integer", "format": "int64"}, "original_name": map[string]any{"type": "string"}, "content_type": map[string]any{"type": "string"},
		"width": map[string]any{"type": "integer"}, "height": map[string]any{"type": "integer"}, "size_bytes": map[string]any{"type": "integer", "format": "int64"},
		"created_at": map[string]any{"type": "string", "format": "date-time"}, "thumbnail_url": map[string]any{"type": "string"}, "original_url": map[string]any{"type": "string"},
	}}
}

func discoveryFeedResponseSchema() map[string]any {
	return map[string]any{"type": "object", "required": []string{"data", "meta"}, "properties": map[string]any{
		"data": map[string]any{"type": "array", "items": map[string]any{"$ref": "#/components/schemas/DiscoveryFeedItem"}},
		"meta": map[string]any{"type": "object", "properties": map[string]any{"page": map[string]any{"type": "integer"}, "per_page": map[string]any{"type": "integer"}, "total": map[string]any{"type": "integer"}}},
	}}
}

func discoverySubmissionResponseSchema() map[string]any {
	return map[string]any{"type": "object", "required": []string{"data"}, "properties": map[string]any{
		"data": map[string]any{"type": "object", "required": []string{"media_id", "status"}, "properties": map[string]any{"media_id": map[string]any{"type": "integer", "format": "int64"}, "status": map[string]any{"type": "string", "enum": []string{"pending"}}}},
	}}
}

func mediaUploadResponseSchema() map[string]any {
	return map[string]any{"type": "object", "required": []string{"data"}, "properties": map[string]any{"data": map[string]any{
		"type": "object", "required": []string{"id", "upload_session_id", "status", "status_url", "original_name", "content_type", "size_bytes", "width", "height", "replayed", "links"},
		"properties": map[string]any{
			"id": map[string]any{"type": "integer", "format": "int64"}, "upload_session_id": map[string]any{"type": "integer", "format": "int64"},
			"status_url": map[string]any{"type": "string", "format": "uri-reference"},
			"status":     map[string]any{"type": "string", "enum": []string{"ready", "processing", "failed"}}, "original_name": map[string]any{"type": "string"},
			"content_type": map[string]any{"type": "string"}, "size_bytes": map[string]any{"type": "integer", "format": "int64"},
			"width": map[string]any{"type": "integer"}, "height": map[string]any{"type": "integer"}, "replayed": map[string]any{"type": "boolean"},
			"error_code": map[string]any{"type": "string"},
			"links":      map[string]any{"nullable": true, "type": "object", "additionalProperties": map[string]any{"type": "string"}},
		},
	}}}
}

func batchMediaUploadResponseSchema() map[string]any {
	return map[string]any{"type": "object", "required": []string{"data"}, "properties": map[string]any{"data": map[string]any{
		"type": "object", "required": []string{"items", "summary"}, "properties": map[string]any{
			"items": map[string]any{"type": "array", "items": map[string]any{"type": "object", "required": []string{"client_name", "status"}, "properties": map[string]any{
				"client_name": map[string]any{"type": "string"}, "status": map[string]any{"type": "string"},
				"error": map[string]any{"type": "object", "nullable": true, "properties": map[string]any{"code": map[string]any{"type": "string"}}},
				"id":    map[string]any{"type": "integer", "format": "int64"}, "upload_session_id": map[string]any{"type": "integer", "format": "int64"},
				"status_url": map[string]any{"type": "string", "format": "uri-reference"}, "original_name": map[string]any{"type": "string"},
				"content_type": map[string]any{"type": "string"}, "size_bytes": map[string]any{"type": "integer", "format": "int64"},
				"width": map[string]any{"type": "integer"}, "height": map[string]any{"type": "integer"}, "replayed": map[string]any{"type": "boolean"},
				"links": map[string]any{"type": "object", "nullable": true, "additionalProperties": map[string]any{"type": "string"}},
			}}},
			"summary": map[string]any{"type": "object", "required": []string{"total", "accepted", "failed"}, "properties": map[string]any{
				"total": map[string]any{"type": "integer"}, "accepted": map[string]any{"type": "integer"}, "failed": map[string]any{"type": "integer"},
			}},
		},
	}}}
}

func collectionSchema() map[string]any {
	return map[string]any{"type": "object", "required": []string{"id", "name"}, "properties": map[string]any{
		"id": map[string]any{"type": "integer", "format": "int64"}, "name": map[string]any{"type": "string"},
		"user_id": map[string]any{"type": "integer", "format": "int64"}, "parent_id": map[string]any{"type": "integer", "format": "int64", "nullable": true},
		"cover_media_id": map[string]any{"type": "integer", "format": "int64", "nullable": true}, "visibility": map[string]any{"type": "string", "enum": []string{"private", "unlisted", "public"}},
		"created_at": map[string]any{"type": "string"}, "updated_at": map[string]any{"type": "string"},
	}}
}

func shareLinkSchema() map[string]any {
	return map[string]any{"type": "object", "required": []string{"id", "media_id", "url", "status", "created_at"}, "properties": map[string]any{
		"id": map[string]any{"type": "integer", "format": "int64"}, "media_id": map[string]any{"type": "integer", "format": "int64"},
		"token": map[string]any{"type": "string", "description": "Returned only once when the share link is created."}, "url": map[string]any{"type": "string"},
		"token_prefix": map[string]any{"type": "string"}, "status": map[string]any{"type": "string", "enum": []string{"active", "revoked"}},
		"expires_at": map[string]any{"type": "string", "format": "date-time", "nullable": true}, "created_at": map[string]any{"type": "string", "format": "date-time"},
	}}
}

func signedMediaURLSchema() map[string]any {
	return map[string]any{"type": "object", "required": []string{"url", "variant", "expires_at"}, "properties": map[string]any{
		"url": map[string]any{"type": "string", "description": "Opaque application-signed delivery URL."}, "variant": map[string]any{"type": "string", "enum": []string{"original", "thumbnail", "medium"}}, "expires_at": map[string]any{"type": "string", "format": "date-time"},
	}}
}

func hotlinkPolicySchema() map[string]any {
	return map[string]any{"type": "object", "required": []string{"media_id", "mode", "allow_no_referer"}, "properties": map[string]any{
		"media_id": map[string]any{"type": "integer", "format": "int64"}, "mode": map[string]any{"type": "string", "enum": []string{"off", "referer", "signed", "hybrid"}}, "allow_no_referer": map[string]any{"type": "boolean"},
	}}
}

func hotlinkDomainSchema() map[string]any {
	return map[string]any{"type": "object", "required": []string{"id", "host", "status"}, "properties": map[string]any{
		"id": map[string]any{"type": "integer", "format": "int64"}, "host": map[string]any{"type": "string"}, "status": map[string]any{"type": "string", "enum": []string{"active", "revoked"}}, "created_at": map[string]any{"type": "string", "format": "date-time", "nullable": true},
	}}
}

func mediaAccessLogSchema() map[string]any {
	return map[string]any{"type": "object", "required": []string{"id", "media_asset_id", "variant", "delivery_mode", "result", "accessed_at"}, "properties": map[string]any{
		"id":             map[string]any{"type": "integer", "format": "int64"},
		"media_asset_id": map[string]any{"type": "integer", "format": "int64"},
		"share_link_id":  map[string]any{"type": "integer", "format": "int64", "nullable": true},
		"variant":        map[string]any{"type": "string", "enum": []string{"original", "thumbnail", "medium"}},
		"delivery_mode":  map[string]any{"type": "string", "enum": []string{"off", "referer", "signed", "hybrid", "share"}},
		"result":         map[string]any{"type": "string", "enum": []string{"allowed", "denied", "not_found", "expired", "invalid"}},
		"referer_host":   map[string]any{"type": "string", "nullable": true},
		"accessed_at":    map[string]any{"type": "string", "format": "date-time"},
	}}
}

func personalAPITokenSchema() map[string]any {
	return map[string]any{"type": "object", "required": []string{"id", "name", "prefix", "scopes", "status", "created_at"}, "properties": map[string]any{
		"id": map[string]any{"type": "integer", "format": "int64"}, "name": map[string]any{"type": "string"},
		"token":  map[string]any{"type": "string", "description": "Returned only once when a token is created or rotated."},
		"prefix": map[string]any{"type": "string"}, "scopes": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
		"status":     map[string]any{"type": "string", "enum": []string{"active", "disabled", "revoked"}},
		"expires_at": map[string]any{"type": "string", "format": "date-time", "nullable": true}, "created_at": map[string]any{"type": "string", "format": "date-time"},
		"last_used_at": map[string]any{"type": "string", "format": "date-time", "nullable": true}, "last_used_ip": map[string]any{"type": "string", "nullable": true},
		"usage_count": map[string]any{"type": "integer", "format": "int64"},
	}}
}

func adSlotSchema() map[string]any {
	return map[string]any{"type": "object", "required": []string{"id", "name", "placement", "creative_type", "creative_content"}, "properties": map[string]any{
		"id": map[string]any{"type": "integer", "format": "int64"}, "name": map[string]any{"type": "string"},
		"placement":        map[string]any{"type": "string", "enum": []string{"header", "footer", "left", "right"}},
		"creative_type":    map[string]any{"type": "string", "enum": []string{"text", "image", "script"}},
		"creative_content": map[string]any{"type": "string"}, "target_url": map[string]any{"type": "string", "format": "uri-reference"},
	}}
}

func adSlotListResponseSchema() map[string]any {
	return map[string]any{"type": "object", "required": []string{"data"}, "properties": map[string]any{
		"data": map[string]any{"type": "array", "items": map[string]any{"$ref": "#/components/schemas/AdSlot"}},
	}}
}

func collectionListOperation(operationID string) map[string]any {
	return map[string]any{"operationId": operationID, "responses": map[string]any{"200": map[string]any{"description": "Own collection list", "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"type": "object", "required": []string{"data"}, "properties": map[string]any{"data": map[string]any{"type": "array", "items": map[string]any{"$ref": "#/components/schemas/Collection"}}}}}}}, "401": errorResponse()}}
}

func collectionWriteOperation(operationID, status string, item ...bool) map[string]any {
	operation := map[string]any{"operationId": operationID, "requestBody": jsonBody("CollectionRequest", map[string]any{"type": "object", "required": []string{"name"}, "properties": map[string]any{"name": map[string]any{"type": "string", "maxLength": 120}, "parent_id": map[string]any{"type": "integer", "nullable": true}, "cover_media_id": map[string]any{"type": "integer", "nullable": true}, "visibility": map[string]any{"type": "string", "enum": []string{"private", "unlisted", "public"}}}}), "responses": map[string]any{status: jsonResponse("Collection"), "401": errorResponse(), "404": errorResponse(), "422": errorResponse()}}
	if len(item) > 0 && item[0] {
		operation["parameters"] = []map[string]any{pathParameter("id")}
	}
	return operation
}

func collectionDeleteOperation(operationID string) map[string]any {
	return map[string]any{"operationId": operationID, "parameters": []map[string]any{pathParameter("id")}, "responses": map[string]any{"204": map[string]any{"description": "Collection deleted"}, "401": errorResponse(), "404": errorResponse()}}
}

func multipartUploadBody() map[string]any {
	return map[string]any{"required": true, "content": map[string]any{"multipart/form-data": map[string]any{
		"schema": map[string]any{"type": "object", "required": []string{"file"}, "properties": map[string]any{"file": map[string]any{"type": "string", "format": "binary"}}},
	}}}
}

func batchMultipartUploadBody() map[string]any {
	return map[string]any{"required": true, "content": map[string]any{"multipart/form-data": map[string]any{
		"schema": map[string]any{"type": "object", "required": []string{"files"}, "properties": map[string]any{
			"files": map[string]any{"type": "array", "minItems": 1, "maxItems": 5, "items": map[string]any{"type": "string", "format": "binary"}},
		}},
	}}}
}

func resourceOptionSchema() map[string]any {
	return map[string]any{"type": "object", "required": []string{"value", "label"}, "properties": map[string]any{"value": map[string]any{"type": "string"}, "label": map[string]any{"type": "string"}}}
}

func resourceFieldSchema() map[string]any {
	return map[string]any{"type": "object", "required": []string{"name", "label", "type", "visible", "readable", "writable", "sensitive"}, "properties": map[string]any{"name": map[string]any{"type": "string"}, "label": map[string]any{"type": "string"}, "type": map[string]any{"type": "string"}, "required": map[string]any{"type": "boolean"}, "visible": map[string]any{"type": "boolean"}, "readable": map[string]any{"type": "boolean"}, "writable": map[string]any{"type": "boolean"}, "sensitive": map[string]any{"type": "boolean"}, "options": map[string]any{"type": "array", "items": map[string]any{"$ref": "#/components/schemas/ResourceOption"}}}}
}

func resourceColumnSchema() map[string]any {
	return map[string]any{"type": "object", "required": []string{"name", "label", "sortable"}, "properties": map[string]any{"name": map[string]any{"type": "string"}, "label": map[string]any{"type": "string"}, "sortable": map[string]any{"type": "boolean"}}}
}

func resourceFilterSchema() map[string]any {
	return map[string]any{"type": "object", "required": []string{"name", "label", "type"}, "properties": map[string]any{"name": map[string]any{"type": "string"}, "label": map[string]any{"type": "string"}, "type": map[string]any{"type": "string", "enum": []string{"select", "multi-select", "boolean", "text", "date-range", "relation"}}, "options": map[string]any{"type": "array", "items": map[string]any{"$ref": "#/components/schemas/ResourceOption"}}, "relation": map[string]any{"type": "string"}}}
}

func resourceActionSchema() map[string]any {
	return map[string]any{"type": "object", "required": []string{"name", "label", "permission", "batch"}, "properties": map[string]any{"name": map[string]any{"type": "string"}, "label": map[string]any{"type": "string"}, "kind": map[string]any{"type": "string"}, "permission": map[string]any{"type": "string"}, "batch": map[string]any{"type": "boolean"}, "payload": map[string]any{"type": "string"}, "payload_fields": map[string]any{"type": "array", "items": map[string]any{"$ref": "#/components/schemas/ActionPayloadField"}}}}
}

func actionPayloadFieldSchema() map[string]any {
	return map[string]any{"type": "object", "required": []string{"name", "label", "type"}, "properties": map[string]any{"name": map[string]any{"type": "string"}, "label": map[string]any{"type": "string"}, "type": map[string]any{"type": "string", "enum": []string{"text", "number", "boolean", "select"}}, "required": map[string]any{"type": "boolean"}, "options": map[string]any{"type": "array", "items": map[string]any{"$ref": "#/components/schemas/ResourceOption"}}}}
}

func auditCleanupRequestSchema() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{
		"mode":           map[string]any{"$ref": "#/components/schemas/AuditCleanupMode"},
		"retention_days": map[string]any{"type": "integer", "minimum": 1, "maximum": 3650},
		"ids":            map[string]any{"type": "array", "items": map[string]any{"type": "integer", "format": "int64"}, "maxItems": 1000},
		"action":         map[string]any{"type": "string"},
		"user_id":        map[string]any{"type": "string"},
		"confirmation":   map[string]any{"type": "string", "description": "Must be DELETE for destructive scopes."},
	}}
}

func auditCleanupResponseSchema() map[string]any {
	return map[string]any{"type": "object", "required": []string{"deleted", "mode"}, "properties": map[string]any{
		"deleted":        map[string]any{"type": "integer"},
		"mode":           map[string]any{"$ref": "#/components/schemas/AuditCleanupMode"},
		"retention_days": map[string]any{"type": "integer"},
		"cutoff":         map[string]any{"type": "string", "format": "date-time", "nullable": true},
	}}
}

func notificationSchema() map[string]any {
	return map[string]any{"type": "object", "required": []string{"id", "user_id", "type", "title", "body", "read_at", "created_at"}, "properties": map[string]any{
		"id": map[string]any{"type": "integer", "format": "int64"}, "user_id": map[string]any{"type": "integer", "format": "int64"}, "type": map[string]any{"type": "string"}, "title": map[string]any{"type": "string"}, "body": map[string]any{"type": "string"}, "url": map[string]any{"type": "string", "nullable": true}, "read_at": map[string]any{"type": "string", "format": "date-time", "nullable": true}, "created_at": map[string]any{"type": "string", "format": "date-time"},
	}}
}

func notificationUnreadCountSchema() map[string]any {
	return map[string]any{"type": "object", "required": []string{"count"}, "properties": map[string]any{"count": map[string]any{"type": "integer"}}}
}

func notificationReadResponseSchema() map[string]any {
	return map[string]any{"type": "object", "required": []string{"id", "read"}, "properties": map[string]any{"id": map[string]any{"type": "integer"}, "read": map[string]any{"type": "boolean"}}}
}

func notificationMarkAllReadResponseSchema() map[string]any {
	return map[string]any{"type": "object", "required": []string{"updated"}, "properties": map[string]any{"updated": map[string]any{"type": "integer"}}}
}

func resourceManifestSchema() map[string]any {
	return map[string]any{"type": "object", "required": []string{"name", "label", "route", "permissions", "navigation", "fields", "columns"}, "properties": map[string]any{"name": map[string]any{"type": "string"}, "label": map[string]any{"type": "string"}, "route": map[string]any{"type": "string"}, "permissions": map[string]any{"type": "array", "items": map[string]any{"type": "string"}}, "navigation": map[string]any{"type": "object", "required": []string{"group", "order"}, "properties": map[string]any{"group": map[string]any{"type": "string"}, "order": map[string]any{"type": "integer"}, "hidden": map[string]any{"type": "boolean"}}}, "data_scope": map[string]any{"$ref": "#/components/schemas/DataScope"}, "owner_field": map[string]any{"type": "string"}, "soft_delete": map[string]any{"type": "boolean"}, "fields": map[string]any{"type": "array", "items": map[string]any{"$ref": "#/components/schemas/ResourceField"}}, "columns": map[string]any{"type": "array", "items": map[string]any{"$ref": "#/components/schemas/ResourceColumn"}}, "actions": map[string]any{"type": "array", "items": map[string]any{"$ref": "#/components/schemas/ResourceAction"}}, "filters": map[string]any{"type": "array", "items": map[string]any{"$ref": "#/components/schemas/ResourceFilter"}}, "relations": map[string]any{"type": "array", "items": map[string]any{"$ref": "#/components/schemas/ResourceRelation"}}, "form_groups": map[string]any{"type": "array", "items": map[string]any{"$ref": "#/components/schemas/ResourceFormGroup"}}, "details": map[string]any{"type": "array", "items": map[string]any{"$ref": "#/components/schemas/ResourceDetailSection"}}, "dependencies": map[string]any{"type": "array", "items": map[string]any{"$ref": "#/components/schemas/ResourceFieldDependency"}}}}
}

func resourceRelationSchema() map[string]any {
	return map[string]any{"type": "object", "required": []string{"name", "kind", "resource", "field", "foreign_field", "label_field", "selectable", "multiple"}, "properties": map[string]any{"name": map[string]any{"type": "string"}, "kind": map[string]any{"type": "string", "enum": []string{"belongsTo", "hasMany"}}, "resource": map[string]any{"type": "string"}, "field": map[string]any{"type": "string"}, "foreign_field": map[string]any{"type": "string"}, "label_field": map[string]any{"type": "string"}, "selectable": map[string]any{"type": "boolean"}, "multiple": map[string]any{"type": "boolean"}, "permission": map[string]any{"type": "string"}, "filter_fields": map[string]any{"type": "array", "items": map[string]any{"type": "string"}}}}
}

func resourceFormGroupSchema() map[string]any {
	return map[string]any{"type": "object", "required": []string{"name", "label", "fields"}, "properties": map[string]any{"name": map[string]any{"type": "string"}, "label": map[string]any{"type": "string"}, "columns": map[string]any{"type": "integer", "minimum": 1, "maximum": 4}, "fields": map[string]any{"type": "array", "items": map[string]any{"type": "string"}}}}
}

func resourceDetailSectionSchema() map[string]any {
	return map[string]any{"type": "object", "required": []string{"name", "label", "fields"}, "properties": map[string]any{"name": map[string]any{"type": "string"}, "label": map[string]any{"type": "string"}, "fields": map[string]any{"type": "array", "items": map[string]any{"type": "string"}}}}
}

func resourceFieldDependencySchema() map[string]any {
	return map[string]any{"type": "object", "required": []string{"field", "on", "value"}, "properties": map[string]any{"field": map[string]any{"type": "string"}, "on": map[string]any{"type": "string"}, "value": map[string]any{"type": "string"}}}
}

func relationOptionSchema() map[string]any {
	return map[string]any{"type": "object", "required": []string{"value", "label"}, "properties": map[string]any{"value": map[string]any{"type": "string"}, "label": map[string]any{"type": "string"}}}
}

func relationOptionListSchema() map[string]any {
	return map[string]any{"type": "object", "required": []string{"data", "meta"}, "properties": map[string]any{
		"data": map[string]any{"type": "array", "items": map[string]any{"$ref": "#/components/schemas/RelationOption"}},
		"meta": map[string]any{"$ref": "#/components/schemas/ResourceListMeta"},
	}}
}

func resourceListMetaSchema() map[string]any {
	return map[string]any{"type": "object", "required": []string{"page", "per_page", "total", "last_page"}, "properties": map[string]any{
		"page": map[string]any{"type": "integer"}, "per_page": map[string]any{"type": "integer"}, "total": map[string]any{"type": "integer"}, "last_page": map[string]any{"type": "integer"},
	}}
}

func actionRequestSchema() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{
		"ids":       map[string]any{"type": "array", "minItems": 1, "maxItems": 1000, "items": map[string]any{"type": "integer", "format": "int64"}},
		"selection": map[string]any{"type": "object", "required": []string{"mode"}, "properties": map[string]any{"mode": map[string]any{"type": "string", "enum": []string{"ids", "query"}}, "ids": map[string]any{"type": "array", "items": map[string]any{"type": "integer", "format": "int64"}}, "query": map[string]any{"type": "object", "additionalProperties": map[string]any{"type": "string"}}, "exclude_ids": map[string]any{"type": "array", "items": map[string]any{"type": "integer", "format": "int64"}}}},
		"payload":   map[string]any{"type": "object", "additionalProperties": true},
	}}
}

func actionResponseSchema() map[string]any {
	return map[string]any{"type": "object", "required": []string{"action", "requested", "succeeded", "failed", "skipped", "failures", "skips"}, "properties": map[string]any{
		"action": map[string]any{"type": "string"}, "requested": map[string]any{"type": "integer"}, "succeeded": map[string]any{"type": "integer"}, "failed": map[string]any{"type": "integer"}, "skipped": map[string]any{"type": "integer"},
		"failures": map[string]any{"type": "array", "items": map[string]any{"type": "object", "required": []string{"id", "code"}, "properties": map[string]any{"id": map[string]any{"type": "integer", "format": "int64"}, "code": map[string]any{"type": "string"}}}},
		"skips":    map[string]any{"type": "array", "items": map[string]any{"type": "object", "required": []string{"id", "code"}, "properties": map[string]any{"id": map[string]any{"type": "integer", "format": "int64"}, "code": map[string]any{"type": "string"}}}},
	}}
}

func globalSearchResultSchema() map[string]any {
	return map[string]any{"type": "object", "required": []string{"resource", "label", "id", "title", "route"}, "properties": map[string]any{
		"resource": map[string]any{"type": "string"}, "label": map[string]any{"type": "string"}, "id": map[string]any{"oneOf": []map[string]any{{"type": "string"}, {"type": "integer"}}}, "title": map[string]any{"type": "string"}, "subtitle": map[string]any{"type": "string"}, "route": map[string]any{"type": "string"},
	}}
}

func globalSearchResponseSchema() map[string]any {
	return map[string]any{"type": "object", "required": []string{"data"}, "properties": map[string]any{
		"data": map[string]any{"type": "array", "items": map[string]any{"$ref": "#/components/schemas/GlobalSearchResult"}},
	}}
}

func jsonBody(name string, schema map[string]any) map[string]any {
	return map[string]any{"required": true, "content": map[string]any{"application/json": map[string]any{"schema": schema, "x-schema-name": name}}}
}

func jsonResponse(name string) map[string]any {
	return map[string]any{"description": "Successful response", "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"type": "object"}, "x-schema-name": name}}}
}

func errorResponse() map[string]any {
	return map[string]any{"description": "Error response", "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"$ref": "#/components/schemas/Error"}}}}
}

func operation(operationID string) map[string]any {
	return map[string]any{"operationId": operationID, "responses": map[string]any{"200": jsonResponse(operationID + "Response"), "401": errorResponse(), "403": errorResponse()}}
}

func listOperation(operationID string, parameters []map[string]any) map[string]any {
	result := operation(operationID)
	result["parameters"] = parameters
	return result
}

func globalSearchOperation() map[string]any {
	return map[string]any{
		"operationId": "globalSearch",
		"parameters":  []map[string]any{queryParameter("q", "string")},
		"responses": map[string]any{
			"200": map[string]any{"description": "Search results", "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"$ref": "#/components/schemas/GlobalSearchResponse"}}}},
			"401": errorResponse(), "403": errorResponse(),
		},
	}
}

func exportOperation() map[string]any {
	return map[string]any{
		"operationId": "exportResource",
		"parameters":  []map[string]any{pathParameter("resource"), queryParameter("search", "string"), queryParameter("status", "string"), queryParameter("sort", "string"), queryParameter("dir", "string")},
		"responses": map[string]any{
			"200": map[string]any{"description": "CSV export", "content": map[string]any{"text/csv": map[string]any{"schema": map[string]any{"type": "string", "format": "binary"}}}},
			"401": errorResponse(), "403": errorResponse(), "404": errorResponse(),
		},
	}
}

func resourceActionOperation() map[string]any {
	return map[string]any{
		"operationId": "executeResourceAction",
		"parameters":  []map[string]any{pathParameter("resource"), pathParameter("action")},
		"requestBody": jsonBody("ActionRequest", actionRequestSchema()),
		"responses": map[string]any{
			"200": map[string]any{"description": "Action executed", "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"type": "object", "properties": map[string]any{"data": map[string]any{"$ref": "#/components/schemas/ActionResponse"}}}}}},
			"401": errorResponse(), "403": errorResponse(), "404": errorResponse(), "422": errorResponse(),
		},
	}
}

func relationOptionsOperation() map[string]any {
	return map[string]any{"operationId": "resourceRelationOptions", "parameters": []map[string]any{pathParameter("resource"), pathParameter("relation"), queryParameter("search", "string"), queryParameter("selected", "string"), queryParameter("page", "integer"), queryParameter("per_page", "integer")}, "responses": map[string]any{"200": map[string]any{"description": "Relation options", "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"$ref": "#/components/schemas/RelationOptionList"}}}}, "401": errorResponse(), "403": errorResponse(), "404": errorResponse()}}
}

func relationRecordsOperation() map[string]any {
	return map[string]any{"operationId": "resourceRelationRecords", "parameters": []map[string]any{pathParameter("resource"), pathParameter("id"), pathParameter("relation")}, "responses": map[string]any{"200": map[string]any{"description": "Related records", "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"type": "object", "properties": map[string]any{"data": map[string]any{"type": "array", "items": map[string]any{"$ref": "#/components/schemas/RelationOption"}}}}}}}, "401": errorResponse(), "403": errorResponse(), "404": errorResponse()}}
}

func queryParameter(name string, valueType string) map[string]any {
	return map[string]any{"name": name, "in": "query", "required": false, "schema": map[string]any{"type": valueType}}
}

func pathParameter(name string) map[string]any {
	return map[string]any{"name": name, "in": "path", "required": true, "schema": map[string]any{"type": "string"}}
}

func resourceItemOperation(operationID string) map[string]any {
	return map[string]any{"operationId": operationID, "parameters": []map[string]any{pathParameter("resource"), pathParameter("id")}, "responses": map[string]any{"200": jsonResponse(operationID + "Response"), "401": errorResponse(), "403": errorResponse(), "404": errorResponse()}}
}

func resourceWriteOperation(operationID, success string, item ...bool) map[string]any {
	parameters := []map[string]any{pathParameter("resource")}
	if len(item) > 0 && item[0] {
		parameters = append(parameters, pathParameter("id"))
	}
	return map[string]any{"operationId": operationID, "parameters": parameters, "requestBody": jsonBody("ResourceRecord", map[string]any{"type": "object", "additionalProperties": true}), "responses": map[string]any{success: jsonResponse(operationID + "Response"), "401": errorResponse(), "403": errorResponse(), "404": errorResponse(), "422": errorResponse()}}
}
