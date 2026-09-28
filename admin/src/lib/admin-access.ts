const ADMIN_ENTRY_PERMISSIONS = [
  'admin.users.view',
  'admin.roles.manage',
  'admin.permissions.manage',
  'admin.plans.view',
  'admin.advertising.view',
  'admin.announcements.view',
  'admin.departments.view',
  'admin.media.view',
  'admin.reports.view',
  'admin.api_tokens.view',
  'admin.media_access_logs.view',
  'admin.tasks.view',
  'admin.orders.view',
  'admin.payment_transactions.view',
  'admin.payment_events.view',
  'admin.refunds.view',
  'admin.content_pages.view',
  'admin.footer_navigation.view',
  'admin.friend_links.view',
  'admin.backups.manage',
  'admin.storage.view',
] as const

export function hasAdminAccess(permissions: string[]): boolean {
  const granted = new Set(permissions)
  // The shell is an application boundary, not a role name check. A scoped
  // administrator such as a moderator or API-token operator must be able to
  // enter the shell; own-scope member permissions are intentionally excluded.
  return ADMIN_ENTRY_PERMISSIONS.some((permission) => granted.has(permission))
}
