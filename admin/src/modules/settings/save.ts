export type SettingDefinition = {
  type: string
  group: string
}

export type SettingUpdate = {
  key: string
  value: string
  value_type: string
  group: string
  description: string
}

export function buildSettingUpdates(
  definitions: Record<string, SettingDefinition>,
  values: Record<string, string | undefined>,
  initialValues: Record<string, string | undefined>,
  descriptionForKey: (key: string) => string,
  changedKeys?: ReadonlySet<string>,
): SettingUpdate[] {
  return Object.entries(definitions).flatMap(([key, definition]) => {
    if (changedKeys && !changedKeys.has(key)) return []
    const value = values[key] ?? ''
    if (value === (initialValues[key] ?? '')) return []

    let nextValue = value
    if (definition.type === 'boolean' && nextValue === '') nextValue = 'false'
    if (definition.type === 'secret' && nextValue.trim() === '') nextValue = '__configured__'

    return [{
      key,
      value: nextValue,
      // `select` is a presentation control. The backend stores the selected
      // value as a normal string; sending `select` violates the settings API
      // value_type contract and makes otherwise valid settings impossible to save.
      value_type: definition.type === 'select' ? 'string' : definition.type,
      group: definition.group,
      description: descriptionForKey(key),
    }]
  })
}
