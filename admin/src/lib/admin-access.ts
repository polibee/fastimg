const ADMIN_ENTRY_PERMISSIONS = [
  'admin.users.view',
  'admin.roles.manage',
  'admin.permissions.manage',
  'admin.plans.view',
  'admin.advertising.view',
  'admin.announcements.view',
  'admin.departments.view',
] as const

export function hasAdminAccess(permissions: string[]): boolean {
  const granted = new Set(permissions)
  return ADMIN_ENTRY_PERMISSIONS.some((permission) => granted.has(permission))
}
