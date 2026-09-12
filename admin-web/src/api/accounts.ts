import { api } from '@/api/admin'

export interface Account {
  id: number
  name: string
  email: string
  access_token: string
  claim_code: string
  claim_limit: number
  claimed_count: number
  created_at: string
  updated_at: string
}

export interface AccountInput {
  name: string
  email: string
  access_token: string
  claim_code: string
  claim_limit?: number
}

export function listAccounts() {
  return api<{ items: Account[] }>('/accounts')
}

export function getAccount(id: number) {
  return api<Account>(`/accounts/${id}`)
}

export function createAccount(body: AccountInput) {
  return api<Account>('/accounts', { method: 'POST', body: JSON.stringify(body) })
}

export function updateAccount(id: number, body: AccountInput) {
  return api<Account>(`/accounts/${id}`, { method: 'PUT', body: JSON.stringify(body) })
}

export function updateClaimLimit(id: number, from: number, to: number) {
  return api<Account>(`/accounts/${id}/claim-limit`, {
    method: 'PUT',
    body: JSON.stringify({ from, to }),
  })
}

export function resetClaimed(id: number, from: number) {
  return api<Account>(`/accounts/${id}/reset-claimed`, {
    method: 'POST',
    body: JSON.stringify({ from }),
  })
}

export function deleteAccount(id: number) {
  return api<{ ok: boolean }>(`/accounts/${id}`, { method: 'DELETE' })
}

export function pollCursorAuth(uuid: string, verifier: string) {
  const q = new URLSearchParams({ uuid, verifier })
  return api<Record<string, unknown>>(`/cursor/auth/poll?${q.toString()}`)
}
