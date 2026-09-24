import { apiFetchEnvelope } from '@/lib/api'
import { apiFetch } from '@/lib/api'

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

export interface MemberPlanPrice {
  id: number
  version: string
  currency: string
  amount_minor: number
  billing_period: 'monthly' | 'yearly'
  trial_days: number
  status: string
}

export interface MemberOrder {
  id: number
  public_order_no: string
  user_id: number
  status: string
  currency: string
  total_amount_minor: number
  expires_at: string | null
  paid_at: string | null
  fulfilled_at: string | null
  created_at: string
}

export async function createMemberOrder(input: { plan_id: number; price_id: number; currency: string; billing_period: string }, token: string) {
  const key = globalThis.crypto?.randomUUID?.() ?? `checkout-${Date.now()}`
  return apiFetch<MemberOrder>('/api/v1/orders', { method: 'POST', headers: { 'Idempotency-Key': key }, body: JSON.stringify(input) }, token)
}

export async function getMemberOrder(id: string, token: string) {
  return apiFetch<MemberOrder>(`/api/v1/orders/${id}`, {}, token)
}

export async function listMemberOrders(token: string) {
  return apiFetchEnvelope<MemberOrder[]>('/api/v1/orders?page=1&per_page=50', {}, token)
}

export async function startMemberPayment(id: string, gatewayCode: string, token: string) {
  const key = globalThis.crypto?.randomUUID?.() ?? `payment-${id}-${Date.now()}`
  return apiFetch<{ id: number; status: string; checkout_url: string; provider_payment_id: string }>(`/api/v1/orders/${id}/payments`, { method: 'POST', headers: { 'Idempotency-Key': key }, body: JSON.stringify({ gateway_code: gatewayCode }) }, token)
}

export async function cancelMemberOrder(id: string, token: string) {
  return apiFetch<void>(`/api/v1/orders/${id}/cancel`, { method: 'POST' }, token)
}

export async function completeFakeMemberPayment(id: string, token: string) {
  return apiFetch<{ event_id: string; accepted: boolean }>(`/api/v1/orders/${id}/payments/fake/succeed`, { method: 'POST' }, token)
}
