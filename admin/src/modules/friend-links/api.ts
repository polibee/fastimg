import { apiFetch, apiFetchEnvelope } from '@/lib/api'

export type FriendLink = { id: number; site_name: string; url: string; logo_url?: string; description?: string; contact_email?: string; status?: 'pending' | 'approved' | 'rejected' | 'hidden'; review_note?: string }
export type FriendLinkList = { data: FriendLink[]; meta?: { page?: number; per_page?: number; total?: number; last_page?: number } }
export type FriendLinkInput = { site_name: string; url: string; logo_url: string; description: string; contact_email: string }
export type FriendLinkPresentation = { eyebrow: string; title: string; description: string; empty: string; submit_title: string; submit_description: string; submitted: string }
export const friendApi = {
  list() { return apiFetch<FriendLink[]>('/api/v1/friend-links') },
  presentation(locale = '') { return apiFetch<FriendLinkPresentation>(`/api/v1/site/friend-links/presentation${locale ? `?locale=${encodeURIComponent(locale)}` : ''}`) },
  submit(input: FriendLinkInput) { return apiFetch<{ id: number; status: string }>('/api/v1/friend-links', { method: 'POST', body: JSON.stringify(input) }) },
  adminList(status = '', token: string, page = 1, perPage = 20) { const params = new URLSearchParams({ page: String(page), per_page: String(perPage) }); if (status) params.set('status', status); return apiFetchEnvelope<FriendLink[]>(`/api/v1/admin/friend-links?${params}`, {}, token) },
  review(id: number, status: 'approved' | 'rejected' | 'hidden', review_note: string, token: string) { return apiFetch<FriendLink>(`/api/v1/admin/friend-links/${id}/review`, { method: 'POST', body: JSON.stringify({ status, review_note }) }, token) },
}
