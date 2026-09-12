/**
 * 管理端 API 辅助方法。
 * - 开发：固定 VITE_ADMIN_ENTRY=/__dev__，页面只开 Vite。
 * - 生产：Hash 路由下 pathname 即真实安全入口。
 */

/** 规范化入口：支持误填完整 URL，最终只保留 pathname。 */
export function normalizeAdminEntry(raw: string): string {
  const v = raw.trim()
  if (!v) {
    return ''
  }
  try {
    if (/^https?:\/\//i.test(v)) {
      return new URL(v).pathname.replace(/\/+$/, '') || '/'
    }
  } catch {
    // 忽略 URL 解析失败，按路径处理
  }
  return v.replace(/\/+$/, '') || '/'
}

export function adminApiBase(): string {
  const fromEnv = (import.meta.env.VITE_ADMIN_ENTRY as string | undefined)?.trim()
  if (fromEnv) {
    return `${normalizeAdminEntry(fromEnv)}/api`
  }
  const cached = sessionStorage.getItem('admin_entry')
  if (cached) {
    return `${normalizeAdminEntry(cached)}/api`
  }
  const path = window.location.pathname.replace(/\/+$/, '') || '/'
  return `${path === '/' ? '' : path}/api`
}

export async function api<T>(path: string, init?: RequestInit): Promise<T> {
  const url = adminApiBase() + path
  const res = await fetch(url, {
    credentials: 'include',
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(init?.headers || {}),
    },
  })
  const data = await res.json().catch(() => ({}))
  if (!res.ok) {
    throw new Error((data as { error?: string }).error || `HTTP ${res.status}`)
  }
  return data as T
}

export interface StatusResp {
  initialized: boolean
  admin_entry: string
  public_base_url: string
}

export function rememberEntry(entry: string) {
  // 开发态始终以环境变量入口为准，避免初始化后被改成 /console 导致 Vite 失联
  if (import.meta.env.DEV && import.meta.env.VITE_ADMIN_ENTRY) {
    sessionStorage.setItem('admin_entry', normalizeAdminEntry(String(import.meta.env.VITE_ADMIN_ENTRY)))
    return
  }
  sessionStorage.setItem('admin_entry', normalizeAdminEntry(entry))
}
