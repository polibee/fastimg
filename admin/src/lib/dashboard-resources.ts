export interface DashboardResource {
  name: string
  label: string
  route: string
  permissions: string[]
  data_scope?: 'all' | 'own'
  dataScope?: 'all' | 'own'
}

export function visibleDashboardResources(resources: DashboardResource[], permissions: string[]) {
  const granted = new Set(permissions)
  return resources.filter((resource) => resource.data_scope !== 'own' && resource.dataScope !== 'own'
    && resource.permissions.some((permission) => granted.has(permission)))
}

export function dashboardResourceRoute(resource: Pick<DashboardResource, 'name' | 'route'>) {
  return adminResourcePath(resource.route, resource.name)
}

export function adminResourcePath(route: string, fallbackName: string): string {
  const normalized = route.replace(/^\/+/, '').replace(/^admin\/+/, '')
  return `/admin/${normalized || fallbackName}`
}
