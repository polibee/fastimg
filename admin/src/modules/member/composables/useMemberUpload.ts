import { onBeforeUnmount, ref } from 'vue'
import { ApiError, apiFetch } from '@/lib/api'
import { useAuthStore } from '@/stores/auth'

export type MemberUploadState = 'uploading' | 'processing' | 'ready' | 'failed'
export type MemberUploadLinkKey = 'url' | 'markdown' | 'html' | 'bbcode'

export interface MemberUploadItem {
  id: string
  file: File
  state: MemberUploadState
  errorKey?: string
  sessionID?: number
  links?: Record<string, string>
}

interface UploadResponse {
  id: number
  upload_session_id: number
  status: string
  error_code?: string
  links?: Record<string, string> | null
}

const requiredLinkKeys = ['url', 'markdown', 'html', 'bbcode'] as const

function normalizeLinks(links?: Record<string, string> | null) {
  if (!links || requiredLinkKeys.some((key) => typeof links[key] !== 'string' || !links[key])) return undefined
  return links
}

const acceptedTypes = ['image/jpeg', 'image/png', 'image/gif']

export function useMemberUpload(onUpdated: () => void = () => {}) {
  const auth = useAuthStore()
  const items = ref<MemberUploadItem[]>([])
  const uploading = ref(false)
  let pollTimer: number | undefined

  function setItem(id: string, update: Partial<MemberUploadItem>) {
    const item = items.value.find((entry) => entry.id === id)
    if (item) Object.assign(item, update)
  }

  async function resolveLinks(result: UploadResponse) {
    const directLinks = normalizeLinks(result.links)
    if (directLinks || !auth.token || !result.id) return directLinks
    try {
      const media = await apiFetch<{ links?: Record<string, string> | null }>(`/api/v1/media/${result.id}`, {}, auth.token)
      return normalizeLinks(media.links)
    } catch {
      return undefined
    }
  }

  async function uploadOne(item: MemberUploadItem) {
    if (!auth.token) {
      setItem(item.id, { state: 'failed', errorKey: 'member.media.errors.uploadFailed' })
      return
    }
    setItem(item.id, { state: 'uploading', errorKey: undefined, links: undefined })
    try {
      const body = new FormData()
      body.append('file', item.file)
      const result = await apiFetch<UploadResponse>('/api/v1/uploads', {
        method: 'POST', headers: { 'Idempotency-Key': crypto.randomUUID() }, body,
      }, auth.token)
      if (result.status === 'processing') {
        setItem(item.id, { state: 'processing', sessionID: result.upload_session_id })
        startStatusPolling()
      } else if (result.status === 'ready') {
        setItem(item.id, { state: 'ready', links: await resolveLinks(result) })
        onUpdated()
      } else {
        setItem(item.id, { state: 'failed', errorKey: 'member.media.errors.uploadFailed' })
      }
    } catch (error) {
      const errorKey = error instanceof ApiError && error.code === 'STORAGE_QUOTA_EXCEEDED'
        ? 'member.media.errors.quotaExceeded'
        : error instanceof ApiError && error.code === 'UPLOAD_FILE_TOO_LARGE'
          ? 'member.media.errors.fileTooLarge'
          : 'member.media.errors.uploadFailed'
      setItem(item.id, { state: 'failed', errorKey })
    }
  }

  async function uploadFiles(fileList: FileList | File[]) {
    if (uploading.value) return
    const files = Array.from(fileList)
    const valid = files.filter((file) => acceptedTypes.includes(file.type))
    const rejected = files.length - valid.length
    if (rejected) items.value.push({ id: crypto.randomUUID(), file: new File([], ''), state: 'failed', errorKey: 'member.media.errors.chooseFormat' })
    uploading.value = true
    try {
      for (const file of valid) {
        const item: MemberUploadItem = { id: crypto.randomUUID(), file, state: 'uploading' }
        items.value.push(item)
        await uploadOne(item)
      }
      if (valid.length) onUpdated()
    } finally {
      uploading.value = false
    }
  }

  async function checkUploadStatuses() {
    if (!auth.token) return
    const processing = items.value.filter((item) => item.state === 'processing' && item.sessionID)
    for (const item of processing) {
      try {
        const result = await apiFetch<UploadResponse>(`/api/v1/uploads/${item.sessionID}`, {}, auth.token)
        if (result.status === 'ready') {
          setItem(item.id, { state: 'ready', links: await resolveLinks(result) })
          onUpdated()
        } else if (result.status === 'failed') {
          setItem(item.id, { state: 'failed', errorKey: 'member.media.errors.processingFailed' })
        }
      } catch {
        setItem(item.id, { errorKey: 'member.media.errors.statusUnavailable' })
      }
    }
    if (!items.value.some((item) => item.state === 'processing') && pollTimer !== undefined) {
      window.clearInterval(pollTimer)
      pollTimer = undefined
    }
  }

  function startStatusPolling() {
    if (pollTimer !== undefined) return
    pollTimer = window.setInterval(() => void checkUploadStatuses(), 3000)
  }

  function retry(item: MemberUploadItem) {
    void uploadOne(item)
  }

  onBeforeUnmount(() => {
    if (pollTimer !== undefined) window.clearInterval(pollTimer)
  })

  return { items, uploading, uploadFiles, retry }
}
