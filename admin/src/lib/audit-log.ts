export type AuditMetadata = Record<string, unknown> | string | null

export function parseAuditMetadata(metadata: AuditMetadata): Record<string, unknown> {
  if (!metadata) return {}
  if (typeof metadata !== 'string') return metadata
  try {
    const parsed = JSON.parse(metadata)
    return parsed && typeof parsed === 'object' && !Array.isArray(parsed) ? parsed as Record<string, unknown> : {}
  } catch {
    return {}
  }
}

export function auditOutcome(metadata: AuditMetadata): 'success' | 'error' | 'unknown' {
  const outcome = parseAuditMetadata(metadata).outcome
  return outcome === 'success' || outcome === 'error' ? outcome : 'unknown'
}

export function formatAuditMetadata(metadata: AuditMetadata) {
  if (!metadata) return ''
  if (typeof metadata !== 'string') return JSON.stringify(metadata, null, 2)
  try {
    return JSON.stringify(JSON.parse(metadata), null, 2)
  } catch {
    return metadata
  }
}
