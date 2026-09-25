export interface ResourceFormField {
  name: string
  label: string
  type: string
  hint?: string
  required?: boolean
  options?: readonly { value: string; label: string }[]
  visible?: boolean
  readable?: boolean
  writable?: boolean
  sensitive?: boolean
}

function valueForField(field: ResourceFormField, value: unknown) {
	if (field.type === 'boolean') return Boolean(value)
	if (field.type === 'number' || field.type === 'integer') return value === null || value === undefined || value === '' ? '' : Number(value)
	if (field.type === 'entitlements') {
		if (value && typeof value === 'object') return value
		if (typeof value === 'string' && value.trim()) {
			try { return JSON.parse(value) }
			catch { return {} }
		}
		return {}
	}
	if (field.type === 'datetime-local' && value) {
		const date = new Date(String(value))
		if (!Number.isNaN(date.getTime())) {
			const pad = (part: number) => String(part).padStart(2, '0')
			return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`
		}
	}
	if (field.name === 'status' && (value === null || value === undefined || value === '')) return 'active'
	return value === null || value === undefined ? '' : String(value)
}

export function createResourceForm(fields: readonly ResourceFormField[], record: Record<string, unknown> = {}) {
  return Object.fromEntries(fields.map((field) => [field.name, valueForField(field, record[field.name])]))
}

export function serializeResourceForm(fields: readonly ResourceFormField[], form: Record<string, unknown>, includeField: (field: ResourceFormField) => boolean = () => true) {
  return Object.fromEntries(fields.filter((field) => field.writable !== false && includeField(field)).map((field) => {
		const value = form[field.name]
		if (field.type === 'number' || field.type === 'integer') return [field.name, value === '' || value === null || value === undefined ? null : Number(value)]
		if (field.type === 'boolean') return [field.name, Boolean(value)]
		if (field.type === 'entitlements') return [field.name, JSON.stringify(value || {})]
		return [field.name, value]
  }))
}
