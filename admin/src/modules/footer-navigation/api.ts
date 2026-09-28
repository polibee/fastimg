import { apiFetch } from '@/lib/api'

export type NavigationItem = { id: number; group_id: number; parent_id?: number | null; label: string; target_type: 'page' | 'route' | 'external'; target_value: string; open_in_new_tab: boolean; sort_order: number; is_enabled: boolean }
export type NavigationGroup = { id: number; title: string; locale: string; sort_order: number; is_enabled: boolean }
export type NavigationGroupResult = { group: NavigationGroup; items: NavigationItem[] }
export type NavigationGroupInput = Omit<NavigationGroup, 'id'>
export type NavigationItemInput = Omit<NavigationItem, 'id'>

export const footerApi = {
  list(token: string) { return apiFetch<NavigationGroupResult[]>('/api/v1/admin/footer-navigation', {}, token) },
  createGroup(input: NavigationGroupInput, token: string) { return apiFetch<NavigationGroup>('/api/v1/admin/footer-navigation/groups', { method: 'POST', body: JSON.stringify(input) }, token) },
  updateGroup(id: number, input: NavigationGroupInput, token: string) { return apiFetch<NavigationGroup>(`/api/v1/admin/footer-navigation/groups/${id}`, { method: 'PUT', body: JSON.stringify(input) }, token) },
  deleteGroup(id: number, token: string) { return apiFetch(`/api/v1/admin/footer-navigation/groups/${id}`, { method: 'DELETE' }, token) },
  createItem(input: NavigationItemInput, token: string) { return apiFetch<NavigationItem>('/api/v1/admin/footer-navigation/items', { method: 'POST', body: JSON.stringify(input) }, token) },
  updateItem(id: number, input: NavigationItemInput, token: string) { return apiFetch<NavigationItem>(`/api/v1/admin/footer-navigation/items/${id}`, { method: 'PUT', body: JSON.stringify(input) }, token) },
  deleteItem(id: number, token: string) { return apiFetch(`/api/v1/admin/footer-navigation/items/${id}`, { method: 'DELETE' }, token) },
}
