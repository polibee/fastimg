import { apiFetch } from '@/lib/api'

export type NavigationLocale = 'zh-CN' | 'en-US' | 'all'
export type NavigationTargetType = 'page' | 'route' | 'external' | 'friends'
export type NavigationItem = { id: number; group_id: number; parent_id?: number | null; label: string; label_zh_cn: string; label_en_us: string; target_type: NavigationTargetType; target_value: string; open_in_new_tab: boolean; sort_order: number; is_enabled: boolean }
export type NavigationGroup = { id: number; title: string; title_zh_cn: string; title_en_us: string; locale: NavigationLocale; sort_order: number; is_enabled: boolean }
export type NavigationGroupResult = { group: NavigationGroup; items: NavigationItem[] }
export type NavigationGroupInput = { title_zh_cn: string; title_en_us: string; locale: NavigationLocale; sort_order: number; is_enabled: boolean }
export type NavigationItemInput = { group_id: number; parent_id?: number | null; label_zh_cn: string; label_en_us: string; target_type: NavigationTargetType; target_value: string; open_in_new_tab: boolean; sort_order: number; is_enabled: boolean }

export const footerApi = {
  list(token: string) { return apiFetch<NavigationGroupResult[]>('/api/v1/admin/footer-navigation', {}, token) },
  createGroup(input: NavigationGroupInput, token: string) { return apiFetch<NavigationGroup>('/api/v1/admin/footer-navigation/groups', { method: 'POST', body: JSON.stringify(input) }, token) },
  updateGroup(id: number, input: NavigationGroupInput, token: string) { return apiFetch<NavigationGroup>(`/api/v1/admin/footer-navigation/groups/${id}`, { method: 'PUT', body: JSON.stringify(input) }, token) },
  deleteGroup(id: number, token: string) { return apiFetch(`/api/v1/admin/footer-navigation/groups/${id}`, { method: 'DELETE' }, token) },
  createItem(input: NavigationItemInput, token: string) { return apiFetch<NavigationItem>('/api/v1/admin/footer-navigation/items', { method: 'POST', body: JSON.stringify(input) }, token) },
  updateItem(id: number, input: NavigationItemInput, token: string) { return apiFetch<NavigationItem>(`/api/v1/admin/footer-navigation/items/${id}`, { method: 'PUT', body: JSON.stringify(input) }, token) },
  deleteItem(id: number, token: string) { return apiFetch(`/api/v1/admin/footer-navigation/items/${id}`, { method: 'DELETE' }, token) },
}

export const footerPublicApi = {
  list(locale = '') { return apiFetch<NavigationGroupResult[]>(`/api/v1/site/footer-navigation${locale ? `?locale=${encodeURIComponent(locale)}` : ''}`) },
}
