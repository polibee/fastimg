package routes

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/support"

	admincontrollers "goravel/app/core/admin/controllers"
	authcontrollers "goravel/app/core/auth/controllers"
	"goravel/app/facades"
	adminmiddleware "goravel/app/http/middleware"
	"goravel/app/modules/admin/registry"
	advertisingcontrollers "goravel/app/modules/advertising/controllers"
	albumscontrollers "goravel/app/modules/albums/controllers"
	backupcontrollers "goravel/app/modules/backups/controllers"
	billingcontrollers "goravel/app/modules/billing/controllers"
	collectioncontrollers "goravel/app/modules/collections/controllers"
	contentcontrollers "goravel/app/modules/content/controllers"
	developercontrollers "goravel/app/modules/developer/controllers"
	discoverycontrollers "goravel/app/modules/discovery/controllers"
	footercontrollers "goravel/app/modules/footer_navigation/controllers"
	friendcontrollers "goravel/app/modules/friend_links/controllers"
	mediacontrollers "goravel/app/modules/media/controllers"
	moderationcontrollers "goravel/app/modules/moderation/controllers"
	plancontrollers "goravel/app/modules/plans/controllers"
	seocontrollers "goravel/app/modules/seo/controllers"
	sharecontrollers "goravel/app/modules/shares/controllers"
	uploadcontrollers "goravel/app/modules/uploads/controllers"
	usercontrollers "goravel/app/modules/users/controllers"
	"goravel/app/openapi"
	collectionservices "goravel/app/services/collections"
)

func Web() {
	facades.Route().Get("/", func(ctx http.Context) http.Response {
		return ctx.Response().View().Make("welcome.tmpl", map[string]any{
			"version": support.Version,
		})
	})

	facades.Route().Static("public", "./public")
	facades.Route().Get("/api/openapi.json", func(ctx http.Context) http.Response {
		return ctx.Response().Json(200, http.Json(openapi.Spec()))
	})
	seoController := seocontrollers.NewPublicController()
	facades.Route().Get("/sitemap.xml", seoController.Sitemap)
	facades.Route().Get("/robots.txt", seoController.Robots)
	facades.Route().Get("/api/v1/site/presentation", seoController.Presentation)
	if facades.Config().GetString("app.env", "production") != "production" {
		facades.Route().Get("/api/docs", func(ctx http.Context) http.Response {
			return ctx.Response().Header("Content-Type", "text/html; charset=utf-8").String(200, openapi.DocsHTML())
		})
	}

	userController := usercontrollers.NewUserController()
	planController := plancontrollers.NewPlanController()
	uploadController := uploadcontrollers.NewUploadController()
	mediaController := mediacontrollers.NewMediaController()
	reportController := moderationcontrollers.NewReportController()
	discoveryController := discoverycontrollers.NewController()
	publicAlbumController := albumscontrollers.NewPublicController()
	contentPublicController := contentcontrollers.NewPublicController()
	footerController := footercontrollers.NewController()
	friendLinkController := friendcontrollers.NewController()
	shareController := sharecontrollers.NewShareController()
	advertisingController := advertisingcontrollers.NewAdvertisingController()
	tokenController := developercontrollers.NewTokenController()
	folderController := collectioncontrollers.NewMemberCollectionController(collectionservices.KindFolder)
	albumController := collectioncontrollers.NewMemberCollectionController(collectionservices.KindAlbum)
	albumMediaController := collectioncontrollers.NewMemberAlbumMediaController()
	billingMemberController := billingcontrollers.NewMemberController()
	billingWebhookController := billingcontrollers.NewWebhookController()
	adminFinanceController := billingcontrollers.NewAdminFinanceController()
	facades.Route().Get("/users", userController.Index)
	facades.Route().Get("/api/v1/plans", planController.Index)
	facades.Route().Get("/api/v1/discovery/status", discoveryController.Status)
	facades.Route().Get("/api/v1/discovery/feed", discoveryController.Feed)
	facades.Route().Get("/api/v1/discovery/media/{id}/content", discoveryController.Content)
	facades.Route().Get("/api/v1/public/albums/{id}", publicAlbumController.Show)
	facades.Route().Get("/api/v1/public/albums/{id}/media/{media_id}/content", publicAlbumController.Content)
	facades.Route().Get("/api/v1/site/pages/{slug}", contentPublicController.Show)
	facades.Route().Get("/api/v1/site/footer-navigation", footerController.Public)
	facades.Route().Get("/api/v1/friend-links", friendLinkController.PublicList)
	facades.Route().Post("/api/v1/friend-links", friendLinkController.Submit)
	// The former member submission endpoint is intentionally retired. Keep a
	// tombstone so old clients receive an explicit error instead of a silent
	// success from the framework's unmatched-route fallback.
	facades.Route().Post("/api/v1/media/{id}/discovery-submit", func(ctx http.Context) http.Response {
		return adminmiddleware.APIError(ctx, 410, "DISCOVERY_SUBMISSIONS_DISABLED")
	})
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberSession()).Post("/api/v1/orders", billingMemberController.CreateOrder)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberSession()).Get("/api/v1/orders", billingMemberController.ListOrders)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberSession()).Get("/api/v1/orders/{id}", billingMemberController.ShowOrder)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberSession()).Post("/api/v1/orders/{id}/payments", billingMemberController.StartPayment)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberSession()).Post("/api/v1/orders/{id}/cancel", billingMemberController.CancelOrder)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberSession()).Post("/api/v1/orders/{id}/payments/fake/succeed", billingMemberController.CompleteFakePayment)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberSession()).Get("/api/v1/payment-gateways", billingMemberController.Gateways)
	facades.Route().Post("/api/v1/payment-gateways/{gateway}/webhook", billingWebhookController.Receive)

	authController := authcontrollers.NewAuthController()
	facades.Route().Get("/api/v1/auth/registration-policy", authController.RegistrationPolicy)
	facades.Route().Post("/api/v1/auth/register", authController.Register)
	facades.Route().Get("/api/v1/auth/verify-email", authController.VerifyEmail)
	facades.Route().Post("/api/v1/auth/resend-verification", authController.ResendVerification)
	facades.Route().Post("/api/v1/auth/login", authController.Login)
	facades.Route().Get("/api/v1/auth/me", authController.Me)
	facades.Route().Post("/api/v1/auth/refresh", authController.Refresh)
	facades.Route().Post("/api/v1/auth/logout", authController.Logout)
	facades.Route().Post("/api/v1/auth/logout-all", authController.LogoutAll)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberSession()).Get("/api/v1/subscription", planController.Subscription)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberSession()).Get("/api/v1/ads", advertisingController.Index)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberSession()).Get("/api/v1/me/usage", planController.Usage)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberSession()).Get("/api/v1/me/usage/ledger", planController.UsageLedger)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberScope("upload:write"), adminmiddleware.UploadBodyLimit()).Post("/api/v1/uploads", uploadController.Create)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberSession(), adminmiddleware.UploadBatchBodyLimit()).Post("/api/v1/uploads/batch", uploadController.Batch)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberSession()).Get("/api/v1/uploads/{id}", uploadController.Status)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberSession()).Post("/api/v1/uploads/{id}/retry", uploadController.Retry)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberScope("media:read")).Get("/api/v1/media", mediaController.Index)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberSession()).Get("/api/v1/me/folders", folderController.Index)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberSession()).Post("/api/v1/me/folders", folderController.Create)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberSession()).Patch("/api/v1/me/folders/{id}", folderController.Update)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberSession()).Delete("/api/v1/me/folders/{id}", folderController.Delete)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberSession()).Get("/api/v1/me/albums", albumController.Index)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberSession()).Post("/api/v1/me/albums", albumController.Create)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberSession()).Patch("/api/v1/me/albums/{id}", albumController.Update)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberSession()).Delete("/api/v1/me/albums/{id}", albumController.Delete)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberSession()).Get("/api/v1/me/albums/{id}/media", albumMediaController.Index)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberSession()).Post("/api/v1/me/albums/{id}/media", albumMediaController.Add)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberSession()).Delete("/api/v1/me/albums/{id}/media", albumMediaController.RemoveBatch)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberSession()).Patch("/api/v1/me/albums/{id}/media/order", albumMediaController.Reorder)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberSession()).Post("/api/v1/me/albums/{id}/media/move", albumMediaController.Move)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberSession()).Delete("/api/v1/me/albums/{id}/media/{media_id}", albumMediaController.Remove)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberScope("media:read")).Get("/api/v1/media/{id}", mediaController.Show)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberScope("media:read")).Get("/api/v1/media/{id}/content", mediaController.Content)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberSession()).Patch("/api/v1/media/{id}/visibility", mediaController.UpdateVisibility)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberScope("media:delete")).Delete("/api/v1/media/{id}", mediaController.Delete)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberSession()).Delete("/api/v1/media/trash", mediaController.EmptyTrash)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberSession()).Delete("/api/v1/media/{id}/permanent", mediaController.PermanentDelete)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberSession()).Post("/api/v1/media/{id}/restore", mediaController.Restore)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberSession()).Patch("/api/v1/media/{id}/folder", mediaController.MoveToFolder)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberSession()).Post("/api/v1/media/{id}/share-links", shareController.Create)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberSession()).Post("/api/v1/media/{id}/reports", reportController.Create)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberSession()).Get("/api/v1/me/reports", reportController.Mine)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberSession()).Post("/api/v1/media/{id}/signed-url", shareController.SignedURL)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberSession()).Get("/api/v1/media/{id}/hotlink-policy", shareController.HotlinkPolicy)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberSession()).Put("/api/v1/media/{id}/hotlink-policy", shareController.UpdateHotlinkPolicy)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberSession()).Get("/api/v1/hotlink-domains", shareController.HotlinkDomains)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberSession()).Post("/api/v1/hotlink-domains", shareController.CreateHotlinkDomain)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberSession()).Delete("/api/v1/hotlink-domains/{id}", shareController.DeleteHotlinkDomain)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberSession()).Get("/api/v1/share-links", shareController.Index)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberSession()).Delete("/api/v1/share-links/{id}", shareController.Revoke)
	// Small compatibility surface for PicGo/ShareX-style clients. These aliases
	// map only to the four Personal API Token capabilities and do not expose the
	// browser-only batch, folder, sharing, billing, or admin endpoints.
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberScope("upload:write"), adminmiddleware.UploadBodyLimit()).Post("/api/upload", uploadController.Create)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberScope("media:read")).Get("/api/images", mediaController.Index)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberScope("media:read")).Get("/api/image/{id}", mediaController.Show)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberScope("media:delete")).Delete("/api/image/{id}", mediaController.Delete)
	facades.Route().Middleware(adminmiddleware.RequireAuthentication()).Get("/api/v1/tokens", tokenController.Index)
	facades.Route().Middleware(adminmiddleware.RequireAuthentication()).Post("/api/v1/tokens", tokenController.Create)
	facades.Route().Middleware(adminmiddleware.RequireAuthentication()).Delete("/api/v1/tokens/{id}", tokenController.Delete)
	facades.Route().Middleware(adminmiddleware.RequireAuthentication()).Post("/api/v1/tokens/{id}/rotate", tokenController.Rotate)
	facades.Route().Get("/s/{token}", shareController.Public)
	facades.Route().Get("/i/{id}", shareController.PublicSigned)

	rbacController := admincontrollers.NewRBACController()
	overviewController := admincontrollers.NewOverviewController()
	statisticsController := admincontrollers.NewStatisticsController()
	auditController := admincontrollers.NewAuditController()
	mediaAccessController := admincontrollers.NewMediaAccessController()
	queueController := admincontrollers.NewQueueController()
	adminAlbumMediaController := albumscontrollers.NewAdminMediaController()
	resourceController := admincontrollers.NewResourceController()
	globalSearchController := admincontrollers.NewGlobalSearchController()
	settingsController := admincontrollers.NewSettingsController()
	contentAdminController := contentcontrollers.NewAdminController()
	storageController := admincontrollers.NewStorageController()
	backupController := backupcontrollers.NewBackupController()
	notificationController := admincontrollers.NewNotificationController()
	facades.Route().Middleware(adminmiddleware.RequireAuthentication()).Get("/api/v1/notifications", notificationController.Index)
	facades.Route().Middleware(adminmiddleware.RequireAuthentication()).Get("/api/v1/notifications/unread-count", notificationController.UnreadCount)
	facades.Route().Middleware(adminmiddleware.RequireAuthentication()).Put("/api/v1/notifications/read-all", notificationController.MarkAllRead)
	facades.Route().Middleware(adminmiddleware.RequireAuthentication()).Put("/api/v1/notifications/{id}/read", notificationController.MarkRead)
	facades.Route().Middleware(adminmiddleware.RequireAnyResourcePermission()).Get("/api/v1/admin/registry", resourceController.Index)
	facades.Route().Middleware(adminmiddleware.RequireAnyResourcePermission()).Get("/api/v1/admin/search", globalSearchController.Index)
	facades.Route().Middleware(adminmiddleware.RequirePermission("admin.users.view")).Get("/api/v1/admin/overview", overviewController.Index)
	facades.Route().Middleware(adminmiddleware.RequirePermission("admin.users.view")).Get("/api/v1/admin/statistics/trends", statisticsController.Trends)
	facades.Route().Middleware(adminmiddleware.RequirePermission("admin.users.manage")).Get("/api/v1/admin/users/{id}/subscription", planController.AdminSubscription)
	facades.Route().Middleware(adminmiddleware.RequirePermission("admin.users.manage")).Put("/api/v1/admin/users/{id}/subscription", planController.UpdateAdminSubscription)
	facades.Route().Middleware(adminmiddleware.RequirePermission("admin.users.view")).Get("/api/v1/admin/audit-logs", auditController.Index)
	facades.Route().Middleware(adminmiddleware.RequirePermission("admin.media_access_logs.view")).Get("/api/v1/admin/media-access-logs", mediaAccessController.Index)
	facades.Route().Middleware(adminmiddleware.RequirePermission(admincontrollers.QueueViewPermission)).Get("/api/v1/admin/tasks", queueController.Index)
	facades.Route().Middleware(adminmiddleware.RequirePermission(admincontrollers.QueueRetryPermission)).Post("/api/v1/admin/tasks/{uuid}/retry", queueController.Retry)
	facades.Route().Middleware(adminmiddleware.RequirePermission("admin.albums.view")).Get("/api/v1/admin/albums/{id}/media", adminAlbumMediaController.Index)
	facades.Route().Middleware(adminmiddleware.RequirePermission("admin.albums.update")).Post("/api/v1/admin/albums/{id}/media", adminAlbumMediaController.Add)
	facades.Route().Middleware(adminmiddleware.RequirePermission("admin.albums.update")).Delete("/api/v1/admin/albums/{id}/media", adminAlbumMediaController.Remove)
	facades.Route().Middleware(adminmiddleware.RequirePermission("admin.media.view")).Get("/api/v1/admin/media/{id}/content", mediaController.AdminContent)
	facades.Route().Middleware(adminmiddleware.RequirePermission("admin.orders.view")).Get("/api/v1/admin/orders", adminFinanceController.Orders)
	facades.Route().Middleware(adminmiddleware.RequirePermission("admin.payment_transactions.view")).Get("/api/v1/admin/payment-transactions", adminFinanceController.Transactions)
	facades.Route().Middleware(adminmiddleware.RequirePermission("admin.payment_events.view")).Get("/api/v1/admin/payment-webhook-events", adminFinanceController.WebhookEvents)
	facades.Route().Middleware(adminmiddleware.RequirePermission("admin.refunds.view")).Get("/api/v1/admin/refunds", adminFinanceController.Refunds)
	facades.Route().Middleware(adminmiddleware.RequirePermission("admin.billing.fulfill")).Post("/api/v1/admin/orders/{id}/retry-fulfillment", adminFinanceController.RetryFulfillment)
	facades.Route().Middleware(adminmiddleware.RequirePermission("admin.settings.manage")).Post("/api/v1/admin/audit-logs/cleanup", auditController.Cleanup)
	facades.Route().Middleware(adminmiddleware.RequirePermission("admin.settings.manage")).Get("/api/v1/admin/settings", settingsController.Index)
	facades.Route().Middleware(adminmiddleware.RequirePermission("admin.settings.manage")).Put("/api/v1/admin/settings/{key}", settingsController.Upsert)
	facades.Route().Middleware(adminmiddleware.RequirePermission("admin.content_pages.view")).Get("/api/v1/admin/content-pages", contentAdminController.Index)
	facades.Route().Middleware(adminmiddleware.RequirePermission("admin.content_pages.manage")).Post("/api/v1/admin/content-pages", contentAdminController.Create)
	facades.Route().Middleware(adminmiddleware.RequirePermission("admin.content_pages.view")).Get("/api/v1/admin/content-pages/{id}", contentAdminController.Show)
	facades.Route().Middleware(adminmiddleware.RequirePermission("admin.content_pages.manage")).Put("/api/v1/admin/content-pages/{id}", contentAdminController.Update)
	facades.Route().Middleware(adminmiddleware.RequirePermission("admin.content_pages.manage")).Post("/api/v1/admin/content-pages/{id}/actions/publish", contentAdminController.Publish)
	facades.Route().Middleware(adminmiddleware.RequirePermission("admin.content_pages.manage")).Post("/api/v1/admin/content-pages/{id}/actions/archive", contentAdminController.Archive)
	facades.Route().Middleware(adminmiddleware.RequirePermission("admin.footer_navigation.view")).Get("/api/v1/admin/footer-navigation", footerController.Index)
	facades.Route().Middleware(adminmiddleware.RequirePermission("admin.footer_navigation.manage")).Post("/api/v1/admin/footer-navigation/groups", footerController.SaveGroup)
	facades.Route().Middleware(adminmiddleware.RequirePermission("admin.footer_navigation.manage")).Put("/api/v1/admin/footer-navigation/groups/{id}", footerController.SaveGroup)
	facades.Route().Middleware(adminmiddleware.RequirePermission("admin.footer_navigation.manage")).Delete("/api/v1/admin/footer-navigation/groups/{id}", footerController.DeleteGroup)
	facades.Route().Middleware(adminmiddleware.RequirePermission("admin.footer_navigation.manage")).Post("/api/v1/admin/footer-navigation/items", footerController.SaveItem)
	facades.Route().Middleware(adminmiddleware.RequirePermission("admin.footer_navigation.manage")).Put("/api/v1/admin/footer-navigation/items/{id}", footerController.SaveItem)
	facades.Route().Middleware(adminmiddleware.RequirePermission("admin.footer_navigation.manage")).Delete("/api/v1/admin/footer-navigation/items/{id}", footerController.DeleteItem)
	facades.Route().Middleware(adminmiddleware.RequirePermission("admin.friend_links.view")).Get("/api/v1/admin/friend-links", friendLinkController.AdminList)
	facades.Route().Middleware(adminmiddleware.RequirePermission("admin.friend_links.moderate")).Post("/api/v1/admin/friend-links/{id}/review", friendLinkController.Review)
	facades.Route().Middleware(adminmiddleware.RequirePermission("admin.storage.view")).Get("/api/v1/admin/storage/connections", storageController.Index)
	facades.Route().Middleware(adminmiddleware.RequirePermission("admin.storage.view")).Get("/api/v1/admin/storage/statistics", storageController.Statistics)
	facades.Route().Middleware(adminmiddleware.RequirePermission("admin.storage.manage")).Put("/api/v1/admin/storage/connections/{provider}", storageController.Update)
	facades.Route().Middleware(adminmiddleware.RequirePermission("admin.storage.manage")).Post("/api/v1/admin/storage/connections/{provider}/test", storageController.Test)
	facades.Route().Middleware(adminmiddleware.RequirePermission(backupcontrollers.BackupManagePermission)).Get("/api/v1/admin/backups", backupController.List)
	facades.Route().Middleware(adminmiddleware.RequirePermission(backupcontrollers.BackupManagePermission)).Post("/api/v1/admin/backups", backupController.Create)
	facades.Route().Middleware(adminmiddleware.RequirePermission(backupcontrollers.BackupManagePermission)).Get("/api/v1/admin/backups/{id}", backupController.Show)
	facades.Route().Middleware(adminmiddleware.RequirePermission(backupcontrollers.BackupDownloadPermission)).Get("/api/v1/admin/backups/{id}/download", backupController.Download)
	facades.Route().Middleware(adminmiddleware.RequirePermission(backupcontrollers.BackupManagePermission)).Delete("/api/v1/admin/backups/{id}", backupController.Delete)
	facades.Route().Middleware(adminmiddleware.RequirePermission(backupcontrollers.BackupManagePermission)).Post("/api/v1/admin/backups/validate", backupController.Validate)
	facades.Route().Middleware(adminmiddleware.RequirePermission(backupcontrollers.BackupManagePermission)).Post("/api/v1/admin/backups/restore", backupController.Restore)
	facades.Route().Middleware(adminmiddleware.RequirePermission("admin.roles.manage")).Put("/api/v1/admin/roles/{id}/permissions", rbacController.ReplaceRolePermissions)
	facades.Route().Middleware(adminmiddleware.RequirePermission("admin.roles.manage")).Get("/api/v1/admin/users/{id}/roles", rbacController.UserRoles)
	facades.Route().Middleware(adminmiddleware.RequirePermission("admin.roles.manage")).Put("/api/v1/admin/users/{id}/roles", rbacController.ReplaceUserRoles)
	facades.Route().Middleware(adminmiddleware.RequireResourcePermission("view")).Get("/api/v1/admin/{resource}", resourceController.List)
	facades.Route().Middleware(adminmiddleware.RequireResourcePermission("view")).Get("/api/v1/admin/{resource}/export", resourceController.Export)
	facades.Route().Middleware(adminmiddleware.RequireAuthentication()).Post("/api/v1/admin/{resource}/actions/{action}", resourceController.Action)
	facades.Route().Middleware(adminmiddleware.RequireAuthentication()).Get("/api/v1/admin/{resource}/relations/{relation}/options", resourceController.RelationOptions)
	facades.Route().Middleware(adminmiddleware.RequireAuthentication()).Get("/api/v1/admin/{resource}/{id}/relations/{relation}", resourceController.RelationRecords)
	facades.Route().Middleware(adminmiddleware.RequireResourcePermission("view")).Get("/api/v1/admin/{resource}/{id}", resourceController.Show)
	facades.Route().Middleware(adminmiddleware.RequireResourcePermission("create")).Post("/api/v1/admin/{resource}", resourceController.Create)
	facades.Route().Middleware(adminmiddleware.RequireResourcePermission("update")).Put("/api/v1/admin/{resource}/{id}", resourceController.Update)
	facades.Route().Middleware(adminmiddleware.RequireResourcePermission("delete")).Delete("/api/v1/admin/{resource}/{id}", resourceController.Delete)
	for _, manifest := range registry.AdminRegistry().All() {
		base := "/api/v1/admin/" + manifest.Name
		facades.Route().Middleware(adminmiddleware.RequireResourcePermission("view")).Get(base+"/export", resourceController.Export)
		facades.Route().Middleware(adminmiddleware.RequireResourcePermission("view")).Get(base+"/{id}", resourceController.Show)
		facades.Route().Middleware(adminmiddleware.RequireResourcePermission("update")).Put(base+"/{id}", resourceController.Update)
		facades.Route().Middleware(adminmiddleware.RequireResourcePermission("delete")).Delete(base+"/{id}", resourceController.Delete)
		facades.Route().Middleware(adminmiddleware.RequireAuthentication()).Post(base+"/actions/{action}", resourceController.Action)
		facades.Route().Middleware(adminmiddleware.RequireAuthentication()).Get(base+"/relations/{relation}/options", resourceController.RelationOptions)
		facades.Route().Middleware(adminmiddleware.RequireAuthentication()).Get(base+"/{id}/relations/{relation}", resourceController.RelationRecords)
	}
}
