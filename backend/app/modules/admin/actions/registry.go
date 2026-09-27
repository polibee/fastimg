package actions

import (
	adminactions "goravel/app/core/admin/actions"
	developeractions "goravel/app/modules/developer/actions"
	useractions "goravel/app/modules/users/actions"
)

var applicationRegistry = func() *adminactions.Registry {
	r := adminactions.NewRegistry()
	_ = r.Register(useractions.NewSetStatusHandler())
	_ = r.Register(developeractions.NewSetStatusHandler())
	_ = r.Register(NewResolveHandler())
	_ = r.Register(NewMediaModerationHandler())
	_ = r.Register(NewMediaPermanentDeleteHandler())
	return r
}()

func Registry() *adminactions.Registry      { return applicationRegistry }
func Register(h adminactions.Handler) error { return applicationRegistry.Register(h) }
