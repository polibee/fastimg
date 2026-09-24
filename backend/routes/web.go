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
	billingcontrollers "goravel/app/modules/billing/controllers"
	collectioncontrollers "goravel/app/modules/collections/controllers"
	developercontrollers "goravel/app/modules/developer/controllers"
	mediacontrollers "goravel/app/modules/media/controllers"
	plancontrollers "goravel/app/modules/plans/controllers"
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
	if facades.Config().GetString("app.env", "production") != "production" {
		facades.Route().Get("/api/docs", func(ctx http.Context) http.Response {
			return ctx.Response().Header("Content-Type", "text/html; charset=utf-8").String(200, openapi.DocsHTML())
		})
	}

	userController := usercontrollers.NewUserController()
	planController := plancontrollers.NewPlanController()
	uploadController := uploadcontrollers.NewUploadController()
	mediaController := mediacontrollers.NewMediaController()
	shareController := sharecontrollers.NewShareController()
	advertisingController := advertisingcontrollers.NewAdvertisingController()
	tokenController := developercontrollers.NewTokenController()
	folderController := collectioncontrollers.NewMemberCollectionController(collectionservices.KindFolder)
	albumController := collectioncontrollers.NewMemberCollectionController(collectionservices.KindAlbum)
	billingMemberController := billingcontrollers.NewMemberController()
	facades.Route().Get("/users", userController.Index)
	facades.Route().Get("/api/v1/plans", planController.Index)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication()).Post("/api/v1/orders", billingMemberController.CreateOrder)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication()).Get("/api/v1/orders", billingMemberController.ListOrders)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication()).Get("/api/v1/orders/{id}", billingMemberController.ShowOrder)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication()).Post("/api/v1/orders/{id}/payments", billingMemberController.StartPayment)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication()).Post("/api/v1/orders/{id}/cancel", billingMemberController.CancelOrder)

	authController := authcontrollers.NewAuthController()
	facades.Route().Post("/api/v1/auth/login", authController.Login)
	facades.Route().Get("/api/v1/auth/me", authController.Me)
	facades.Route().Post("/api/v1/auth/refresh", authController.Refresh)
	facades.Route().Post("/api/v1/auth/logout", authController.Logout)
	facades.Route().Post("/api/v1/auth/logout-all", authController.LogoutAll)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication()).Get("/api/v1/subscription", planController.Subscription)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication()).Get("/api/v1/ads", advertisingController.Index)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberScope("usage:read")).Get("/api/v1/me/usage", planController.Usage)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberScope("usage:read")).Get("/api/v1/me/usage/ledger", planController.UsageLedger)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberScope("upload:write"), adminmiddleware.UploadBodyLimit()).Post("/api/v1/uploads", uploadController.Create)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberScope("upload:write"), adminmiddleware.UploadBatchBodyLimit()).Post("/api/v1/uploads/batch", uploadController.Batch)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberScope("links:read")).Get("/api/v1/uploads/{id}", uploadController.Status)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberScope("upload:write")).Post("/api/v1/uploads/{id}/retry", uploadController.Retry)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberScope("media:read")).Get("/api/v1/media", mediaController.Index)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication()).Get("/api/v1/me/folders", folderController.Index)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication()).Post("/api/v1/me/folders", folderController.Create)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication()).Patch("/api/v1/me/folders/{id}", folderController.Update)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication()).Delete("/api/v1/me/folders/{id}", folderController.Delete)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication()).Get("/api/v1/me/albums", albumController.Index)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication()).Post("/api/v1/me/albums", albumController.Create)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication()).Patch("/api/v1/me/albums/{id}", albumController.Update)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication()).Delete("/api/v1/me/albums/{id}", albumController.Delete)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberScope("links:read")).Get("/api/v1/media/{id}", mediaController.Show)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberScope("links:read")).Get("/api/v1/media/{id}/content", mediaController.Content)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberScope("media:delete")).Delete("/api/v1/media/{id}", mediaController.Delete)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberScope("media:delete")).Delete("/api/v1/media/{id}/permanent", mediaController.PermanentDelete)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberScope("media:delete")).Post("/api/v1/media/{id}/restore", mediaController.Restore)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication()).Patch("/api/v1/media/{id}/folder", mediaController.MoveToFolder)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberScope("links:read")).Post("/api/v1/media/{id}/share-links", shareController.Create)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberScope("links:read")).Post("/api/v1/media/{id}/signed-url", shareController.SignedURL)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberScope("links:read")).Get("/api/v1/media/{id}/hotlink-policy", shareController.HotlinkPolicy)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberScope("links:read")).Put("/api/v1/media/{id}/hotlink-policy", shareController.UpdateHotlinkPolicy)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberScope("links:read")).Get("/api/v1/hotlink-domains", shareController.HotlinkDomains)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberScope("links:read")).Post("/api/v1/hotlink-domains", shareController.CreateHotlinkDomain)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberScope("links:read")).Delete("/api/v1/hotlink-domains/{id}", shareController.DeleteHotlinkDomain)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberScope("links:read")).Get("/api/v1/share-links", shareController.Index)
	facades.Route().Middleware(adminmiddleware.RequireMemberAuthentication(), adminmiddleware.RequireMemberScope("links:read")).Delete("/api/v1/share-links/{id}", shareController.Revoke)
	facades.Route().Middleware(adminmiddleware.RequireAuthentication()).Get("/api/v1/tokens", tokenController.Index)
	facades.Route().Middleware(adminmiddleware.RequireAuthentication()).Post("/api/v1/tokens", tokenController.Create)
	facades.Route().Middleware(adminmiddleware.RequireAuthentication()).Delete("/api/v1/tokens/{id}", tokenController.Revoke)
	facades.Route().Middleware(adminmiddleware.RequireAuthentication()).Post("/api/v1/tokens/{id}/rotate", tokenController.Rotate)
	facades.Route().Get("/s/{token}", shareController.Public)
	facades.Route().Get("/i/{id}", shareController.PublicSigned)

	rbacController := admincontrollers.NewRBACController()
	overviewController := admincontrollers.NewOverviewController()
	auditController := admincontrollers.NewAuditController()
	mediaAccessController := admincontrollers.NewMediaAccessController()
	resourceController := admincontrollers.NewResourceController()
	globalSearchController := admincontrollers.NewGlobalSearchController()
	settingsController := admincontrollers.NewSettingsController()
	notificationController := admincontrollers.NewNotificationController()
	facades.Route().Middleware(adminmiddleware.RequireAuthentication()).Get("/api/v1/notifications", notificationController.Index)
	facades.Route().Middleware(adminmiddleware.RequireAuthentication()).Get("/api/v1/notifications/unread-count", notificationController.UnreadCount)
	facades.Route().Middleware(adminmiddleware.RequireAuthentication()).Put("/api/v1/notifications/read-all", notificationController.MarkAllRead)
	facades.Route().Middleware(adminmiddleware.RequireAuthentication()).Put("/api/v1/notifications/{id}/read", notificationController.MarkRead)
	facades.Route().Middleware(adminmiddleware.RequireAnyResourcePermission()).Get("/api/v1/admin/registry", resourceController.Index)
	facades.Route().Middleware(adminmiddleware.RequireAnyResourcePermission()).Get("/api/v1/admin/search", globalSearchController.Index)
	facades.Route().Middleware(adminmiddleware.RequirePermission("admin.users.view")).Get("/api/v1/admin/overview", overviewController.Index)
	facades.Route().Middleware(adminmiddleware.RequirePermission("admin.users.view")).Get("/api/v1/admin/audit-logs", auditController.Index)
	facades.Route().Middleware(adminmiddleware.RequirePermission("admin.media_access_logs.view")).Get("/api/v1/admin/media-access-logs", mediaAccessController.Index)
	facades.Route().Middleware(adminmiddleware.RequirePermission("admin.settings.manage")).Post("/api/v1/admin/audit-logs/cleanup", auditController.Cleanup)
	facades.Route().Middleware(adminmiddleware.RequirePermission("admin.settings.manage")).Get("/api/v1/admin/settings", settingsController.Index)
	facades.Route().Middleware(adminmiddleware.RequirePermission("admin.settings.manage")).Put("/api/v1/admin/settings/{key}", settingsController.Upsert)
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
