export function canDeleteResource(resource: string, record: Record<string, unknown>) {
  return Boolean(record.id) && !(resource === 'roles' && record.name === 'super-admin')
}

export interface ResourceActionLike { name: string; label: string; kind: string; permission: string; batch: boolean; payload?: string; payload_fields?: Array<{ name: string; label: string; type: 'text' | 'number' | 'boolean' | 'select'; required?: boolean; options?: Array<{ value: string; label: string }> }> }

const supportedBatchActionKinds = new Set([
  'archive',
  'user-status',
  'api-token-status',
  'media-moderation',
  'media-permanent-delete',
  'report-resolve',
])

export function executableBatchActions(actions: ResourceActionLike[] | undefined, can: (permission: string) => boolean) {
  return (actions || []).filter((action) => action.batch === true && supportedBatchActionKinds.has(action.kind) && Boolean(action.payload) && can(action.permission))
}
