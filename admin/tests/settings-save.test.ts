import assert from 'node:assert/strict'
import test from 'node:test'
import { buildSettingUpdates } from '../src/modules/settings/save.ts'

const definitions = {
  'site_title': { type: 'string', group: 'seo' },
  'email.provider': { type: 'select', group: 'email' },
  'payment.paypal.client_secret': { type: 'secret', group: 'gateway' },
  'stats.enabled': { type: 'boolean', group: 'statistics' },
} as const

test('does not submit unchanged configured secrets', () => {
  const updates = buildSettingUpdates(
    definitions,
    {
      'site_title': 'FastImg',
      'email.provider': 'smtp',
      'payment.paypal.client_secret': '__configured__',
      'stats.enabled': 'false',
    },
    {
      'site_title': 'FastImg',
      'email.provider': 'smtp',
      'payment.paypal.client_secret': '__configured__',
      'stats.enabled': 'false',
    },
    (key) => key,
  )

  assert.deepEqual(updates, [])
})

test('submits changed values and preserves an intentionally blank secret', () => {
  const updates = buildSettingUpdates(
    definitions,
    {
      'site_title': 'FastImg production',
      'email.provider': 'resend',
      'payment.paypal.client_secret': '',
      'stats.enabled': 'true',
    },
    {
      'site_title': 'FastImg',
      'payment.paypal.client_secret': '__configured__',
      'stats.enabled': 'false',
    },
    (key) => key,
  )

  assert.deepEqual(updates, [
    { key: 'site_title', value: 'FastImg production', value_type: 'string', group: 'seo', description: 'site_title' },
    { key: 'email.provider', value: 'resend', value_type: 'string', group: 'email', description: 'email.provider' },
    { key: 'payment.paypal.client_secret', value: '__configured__', value_type: 'secret', group: 'gateway', description: 'payment.paypal.client_secret' },
    { key: 'stats.enabled', value: 'true', value_type: 'boolean', group: 'statistics', description: 'stats.enabled' },
  ])
})

test('submits only fields explicitly changed by the administrator', () => {
  const updates = buildSettingUpdates(
    definitions,
    {
      site_title: 'FastImg production',
      'email.provider': 'smtp',
      'payment.paypal.client_secret': '__configured__',
      'stats.enabled': 'false',
    },
    {
      site_title: '',
      'email.provider': '',
      'payment.paypal.client_secret': '',
      'stats.enabled': '',
    },
    (key) => key,
    new Set(['site_title']),
  )

  assert.deepEqual(updates, [
    { key: 'site_title', value: 'FastImg production', value_type: 'string', group: 'seo', description: 'site_title' },
  ])
})
