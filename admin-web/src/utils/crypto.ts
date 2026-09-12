const ALPHABET = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'

/** 密码学安全随机串：默认 32 位 A-Za-z0-9（上号码）。 */
export function randomClaimCode(length = 32): string {
  const bytes = new Uint8Array(length)
  crypto.getRandomValues(bytes)
  let out = ''
  for (let i = 0; i < length; i++) {
    out += ALPHABET[bytes[i]! % ALPHABET.length]
  }
  return out
}

function b64url(buf: ArrayBuffer | Uint8Array): string {
  const bytes = buf instanceof Uint8Array ? buf : new Uint8Array(buf)
  let s = ''
  bytes.forEach((b) => {
    s += String.fromCharCode(b)
  })
  return btoa(s).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/g, '')
}

/** PKCE：对齐 CursorTokenTool 网页流程。 */
export async function generatePKCEPair(): Promise<{ codeVerifier: string, codeChallenge: string }> {
  const array = new Uint8Array(32)
  crypto.getRandomValues(array)
  const codeVerifier = b64url(array).substring(0, 43)
  const hash = await crypto.subtle.digest('SHA-256', new TextEncoder().encode(codeVerifier))
  const codeChallenge = b64url(hash)
  return { codeVerifier, codeChallenge }
}

export function generateUUID(): string {
  if (typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID()
  }
  const bytes = new Uint8Array(16)
  crypto.getRandomValues(bytes)
  bytes[6] = (bytes[6]! & 0x0F) | 0x40
  bytes[8] = (bytes[8]! & 0x3F) | 0x80
  const hex = [...bytes].map(b => b.toString(16).padStart(2, '0')).join('')
  return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(12, 16)}-${hex.slice(16, 20)}-${hex.slice(20)}`
}
