import { apiFetch } from '@/lib/api'

export type FriendLink = { id: number; site_name: string; url: string; logo_url?: string; description?: string; contact_email?: string; status?: 'pending' | 'approved' | 'rejected' | 'hidden'; review_note?: string }
export type FriendLinkInput = { site_name: string; url: string; logo_url: string; description: string; contact_email: string }
export const friendApi = {
  list() { return apiFetch<FriendLink[]>('/api/v1/friend-links') },
  submit(input: FriendLinkInput) { return apiFetch<{ id: number; status: string }>('/api/v1/friend-links', { method: 'POST', body: JSON.stringify(input) }) },
  adminList(status = '', token: string) { return apiFetch<FriendLink[]>(`/api/v1/admin/friend-links${status ? `?status=${encodeURIComponent(status)}` : ''}`, {}, token) },
  review(id: number, status: 'approved' | 'rejected' | 'hidden', review_note: string, token: string) { return apiFetch<FriendLink>(`/api/v1/admin/friend-links/${id}/review`, { method: 'POST', body: JSON.stringify({ status, review_note }) }, token) },
}
