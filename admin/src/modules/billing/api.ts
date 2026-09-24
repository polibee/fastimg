import { apiFetchEnvelope } from '@/lib/api'

export type BillingAdminKind = 'orders' | 'transactions' | 'events' | 'refunds'

const endpoints: Record<BillingAdminKind, string> = {
  orders: '/api/v1/admin/orders',
  transactions: '/api/v1/admin/payment-transactions',
  events: '/api/v1/admin/payment-webhook-events',
  refunds: '/api/v1/admin/refunds',
}

export interface BillingAdminList {
  data: Record<string, unknown>[]
  meta?: { page?: number; per_page?: number; total?: number }
}

export function listBillingAdmin(kind: BillingAdminKind, token: string, page = 1, perPage = 20) {
  return apiFetchEnvelope<Record<string, unknown>[]>(`${endpoints[kind]}?page=${page}&per_page=${perPage}`, {}, token)
}
