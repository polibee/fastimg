export type SettingFieldType = 'string' | 'integer' | 'secret' | 'boolean' | 'select' | 'textarea'

export type SettingFieldDefinition = {
  key: string
  type: SettingFieldType
  multiline?: boolean
  span?: 'full'
  visible?: (values: Record<string, string>) => boolean
}

export type SettingGroupDefinition = {
  key: 'site' | 'auth' | 'email' | 'subscription' | 'media' | 'seo' | 'gateway' | 'statistics' | 'code' | 'other'
  fields: SettingFieldDefinition[]
}

const field = (key: string, type: SettingFieldType = 'string', options: Omit<SettingFieldDefinition, 'key' | 'type'> = {}): SettingFieldDefinition => ({ key, type, ...options })

const emailProviderField = (key: string, type: SettingFieldType = 'string'): SettingFieldDefinition => ({
  ...field(key, type),
  visible: (values) => values['email.provider'] === key.split('.')[1],
})

export const settingGroups: SettingGroupDefinition[] = [
  {
    key: 'site',
    fields: [field('site_title'), field('site_url')],
  },
  {
    key: 'auth',
    fields: [
      field('auth.registration.enabled', 'boolean'),
      field('auth.turnstile.enabled', 'boolean'),
      field('auth.turnstile.site_key'),
      field('auth.turnstile.secret_key', 'secret'),
      field('auth.login.turnstile_enabled', 'boolean'),
      field('auth.registration.turnstile_enabled', 'boolean'),
      field('auth.registration.email_verification_enabled', 'boolean'),
      field('auth.registration.email_whitelist_enabled', 'boolean'),
      field('auth.registration.email_whitelist_domains', 'textarea', { span: 'full' }),
      field('auth.registration.verification_expiry_minutes', 'integer'),
      field('auth.registration.verification_resend_protection_enabled', 'boolean'),
      field('auth.registration.verification_resend_email_cooldown_seconds', 'integer'),
      field('auth.registration.verification_resend_ip_cooldown_seconds', 'integer'),
      field('auth.registration.verification_resend_daily_email_limit', 'integer'),
      field('auth.registration.verification_resend_daily_ip_limit', 'integer'),
    ],
  },
  {
    key: 'email',
    fields: [
      field('email.enabled', 'boolean'),
      field('email.provider', 'select'),
      field('email.from_address'),
      field('email.from_name'),
      field('email.reply_to'),
      emailProviderField('email.smtp.host'),
      emailProviderField('email.smtp.port', 'integer'),
      emailProviderField('email.smtp.username'),
      emailProviderField('email.smtp.password', 'secret'),
      emailProviderField('email.smtp.encryption', 'select'),
      emailProviderField('email.aliyun.endpoint'),
      emailProviderField('email.aliyun.access_key_id', 'secret'),
      emailProviderField('email.aliyun.access_key_secret', 'secret'),
      emailProviderField('email.aliyun.account_name'),
      emailProviderField('email.aliyun.reply_to_address', 'boolean'),
      emailProviderField('email.resend.endpoint'),
      emailProviderField('email.resend.api_key', 'secret'),
    ],
  },
  {
    key: 'subscription',
    fields: [
      field('subscription.expiry.email_enabled', 'boolean'),
      field('subscription.expiry.fallback_plan_code'),
      field('subscription.expiry.grace_period_days', 'integer'),
      field('subscription.expiry.reminder_days'),
      field('subscription.expiry.over_quota_policy', 'select'),
    ],
  },
  {
    key: 'media',
    fields: [
      field('upload.max_file_mb', 'integer'),
      field('watermark.text'),
      field('watermark.domain'),
      field('watermark.fallback_image_url'),
    ],
  },
  {
    key: 'seo',
    fields: [
      field('site_description', 'textarea', { span: 'full' }),
      field('site_keywords'),
      field('robots', 'textarea', { span: 'full' }),
      field('sitemap.enabled', 'boolean'),
      field('sitemap.extra_paths', 'textarea', { span: 'full' }),
      field('discover.enabled', 'boolean'),
    ],
  },
  {
    key: 'gateway',
    fields: [
      field('payment.default_gateway', 'select'),
      field('payment.fake.enabled', 'boolean'),
      field('payment.paypal.enabled', 'boolean'),
      field('payment.paypal.environment', 'select'),
      field('payment.paypal.client_id', 'secret'),
      field('payment.paypal.client_secret', 'secret'),
      field('payment.paypal.webhook_id', 'secret'),
      field('payment.paypal.base_url'),
      field('payment.paypal.webhook_url'),
      field('payment.paypal.return_url'),
      field('payment.paypal.cancel_url'),
      field('payment.xcash.enabled', 'boolean'),
      field('payment.xcash.app_id', 'secret'),
      field('payment.xcash.hmac_key', 'secret'),
      field('payment.xcash.base_url'),
      field('payment.xcash.callback_url'),
      field('payment.xcash.return_url'),
      field('payment.nowpayments.enabled', 'boolean'),
      field('payment.nowpayments.api_key', 'secret'),
      field('payment.nowpayments.ipn_secret', 'secret'),
      field('payment.nowpayments.base_url'),
      field('payment.nowpayments.callback_url'),
      field('payment.nowpayments.success_url'),
      field('payment.nowpayments.cancel_url'),
    ],
  },
  {
    key: 'statistics',
    fields: [field('stats.enabled', 'boolean'), field('stats.retention_days', 'integer')],
  },
  {
    key: 'code',
    fields: [
      field('custom.site_verification', 'textarea', { span: 'full' }),
      field('custom.ad_verification', 'textarea', { span: 'full' }),
      field('custom.head', 'textarea', { span: 'full' }),
      field('custom.body', 'textarea', { span: 'full' }),
    ],
  },
  { key: 'other', fields: [] },
]

export const settingDefaults: Record<string, string> = {
  'email.provider': 'smtp',
  'email.from_address': 'noreply@example.com',
  'email.from_name': 'FastImg',
  'email.smtp.encryption': 'starttls',
  'email.smtp.port': '587',
  'email.aliyun.endpoint': 'https://dm.aliyuncs.com',
  'email.resend.endpoint': 'https://api.resend.com',
  'auth.registration.verification_resend_protection_enabled': 'true',
  'auth.registration.verification_resend_email_cooldown_seconds': '60',
  'auth.registration.verification_resend_ip_cooldown_seconds': '10',
  'auth.registration.verification_resend_daily_email_limit': '5',
  'auth.registration.verification_resend_daily_ip_limit': '20',
  'subscription.expiry.email_enabled': 'true',
  'subscription.expiry.fallback_plan_code': 'free',
  'subscription.expiry.grace_period_days': '3',
  'subscription.expiry.reminder_days': '7,3,1',
  'subscription.expiry.over_quota_policy': 'keep_data_block_upload',
}

export const emailProviderOptions = [
  { value: 'smtp', labelKey: 'settings.emailProviders.smtp' },
  { value: 'aliyun', labelKey: 'settings.emailProviders.aliyun' },
  { value: 'resend', labelKey: 'settings.emailProviders.resend' },
]

export const smtpEncryptionOptions = [
  { value: 'starttls', labelKey: 'settings.smtpEncryption.starttls' },
  { value: 'ssl', labelKey: 'settings.smtpEncryption.ssl' },
  { value: 'none', labelKey: 'settings.smtpEncryption.none' },
]

export const paymentEnvironmentOptions = [
  { value: 'sandbox', labelKey: 'settings.paypalEnvironments.sandbox' },
  { value: 'production', labelKey: 'settings.paypalEnvironments.production' },
]

export function settingDefinitions(): Record<string, SettingFieldDefinition & { group: string }> {
  return Object.fromEntries(settingGroups.flatMap((group) => group.fields.map((definition) => [definition.key, { ...definition, group: group.key }])))
}
