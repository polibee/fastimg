const APP_ORIGIN = 'https://app.invalid'

export function resolveLoginRedirect(value: unknown): string {
  if (typeof value !== 'string' || !value.startsWith('/') || value.startsWith('//') || value.includes('\\')) {
    return '/'
  }

  try {
    const url = new URL(value, APP_ORIGIN)
    return url.origin === APP_ORIGIN ? value : '/'
  } catch {
    return '/'
  }
}
