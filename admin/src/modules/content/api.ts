import type { JSONContent } from '@tiptap/core'
import { apiFetch, apiFetchEnvelope } from '@/lib/api'

export type SitePage = {
  id: number
  slug: string
  title: string
  content_json: JSONContent
  excerpt: string
  seo_title: string
  seo_description: string
  status: 'draft' | 'published' | 'archived'
  published_at?: string | null
  created_by?: number
  updated_by?: number
}

export type SitePageInput = {
  slug: string
  title: string
  content_json: JSONContent
  excerpt: string
  seo_title: string
  seo_description: string
}

export const contentApi = {
  list(token: string, page = 1, perPage = 20) {
    return apiFetchEnvelope<SitePage[]>(`/api/v1/admin/content-pages?page=${page}&per_page=${perPage}`, {}, token)
  },
  show(id: string, token: string) {
    return apiFetch<SitePage>(`/api/v1/admin/content-pages/${encodeURIComponent(id)}`, {}, token)
  },
  create(input: SitePageInput, token: string) {
    return apiFetch<SitePage>('/api/v1/admin/content-pages', { method: 'POST', body: JSON.stringify(input) }, token)
  },
  update(id: string, input: SitePageInput, token: string) {
    return apiFetch<SitePage>(`/api/v1/admin/content-pages/${encodeURIComponent(id)}`, { method: 'PUT', body: JSON.stringify(input) }, token)
  },
  publish(id: number, token: string) {
    return apiFetch<SitePage>(`/api/v1/admin/content-pages/${id}/actions/publish`, { method: 'POST', body: JSON.stringify({}) }, token)
  },
  archive(id: number, token: string) {
    return apiFetch<SitePage>(`/api/v1/admin/content-pages/${id}/actions/archive`, { method: 'POST', body: JSON.stringify({}) }, token)
  },
}
