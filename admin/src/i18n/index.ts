import { createI18n } from 'vue-i18n'
import enUSCore from '@/locales/en-US/core.json'
import enUSAuth from '@/locales/en-US/auth.json'
import zhCNCcore from '@/locales/zh-CN/core.json'
import zhCNAuth from '@/locales/zh-CN/auth.json'
import enUSStates from '@/locales/en-US/states.json'
import enUSRbac from '@/locales/en-US/rbac.json'
import zhCNStates from '@/locales/zh-CN/states.json'
import zhCNRbac from '@/locales/zh-CN/rbac.json'
import enUSErrors from '@/locales/en-US/errors.json'
import zhCNErrors from '@/locales/zh-CN/errors.json'
import enUSResource from '@/locales/en-US/resource.json'
import zhCNResource from '@/locales/zh-CN/resource.json'
import enUSMedia from '@/locales/en-US/media.json'
import zhCNMedia from '@/locales/zh-CN/media.json'
import enUSMember from '@/locales/en-US/member.json'
import zhCNMember from '@/locales/zh-CN/member.json'
import enUSBilling from '@/locales/en-US/billing.json'
import zhCNBilling from '@/locales/zh-CN/billing.json'
import enUSSettings from '@/locales/en-US/settings.json'
import zhCNSettings from '@/locales/zh-CN/settings.json'
import enUSStatistics from '@/locales/en-US/statistics.json'
import zhCNStatistics from '@/locales/zh-CN/statistics.json'
import enUSAlbums from '@/locales/en-US/albums.json'
import zhCNAlbums from '@/locales/zh-CN/albums.json'
import enUSStorage from '@/locales/en-US/storage.json'
import zhCNStorage from '@/locales/zh-CN/storage.json'
import enUSTasks from '@/locales/en-US/tasks.json'
import zhCNTasks from '@/locales/zh-CN/tasks.json'
import enUSContent from '@/locales/en-US/content.json'
import zhCNContent from '@/locales/zh-CN/content.json'

export const SUPPORTED_LOCALES = ['zh-CN', 'en-US'] as const

const initialLocale = localStorage.getItem('locale') === 'en-US' ? 'en-US' : 'zh-CN'

const i18n = createI18n({
  legacy: false,
  locale: initialLocale,
  fallbackLocale: 'en-US',
  messages: {
    'zh-CN': { core: zhCNCcore, auth: zhCNAuth, errors: zhCNErrors, states: zhCNStates, rbac: zhCNRbac, resource: zhCNResource, media: zhCNMedia, member: zhCNMember, billing: zhCNBilling, settings: zhCNSettings, statistics: zhCNStatistics, albums: zhCNAlbums, storage: zhCNStorage, tasks: zhCNTasks, content: zhCNContent },
    'en-US': { core: enUSCore, auth: enUSAuth, errors: enUSErrors, states: enUSStates, rbac: enUSRbac, resource: enUSResource, media: enUSMedia, member: enUSMember, billing: enUSBilling, settings: enUSSettings, statistics: enUSStatistics, albums: enUSAlbums, storage: enUSStorage, tasks: enUSTasks, content: enUSContent },
  },
})

export default i18n
