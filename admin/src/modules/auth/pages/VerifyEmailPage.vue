<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { ApiError, errorMessageKey } from '@/lib/api'
import { generatedApi } from '@/generated/api'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const loading = ref(true)
const verified = ref(false)
const pending = ref(route.query.pending === '1')
const error = ref('')

onMounted(async () => {
  const token = typeof route.query.token === 'string' ? route.query.token : ''
  if (!token) {
    loading.value = false
    if (!pending.value) error.value = t('auth.verificationLinkInvalid')
    return
  }
  try {
    await generatedApi.verifyEmail(token)
    verified.value = true
  } catch (cause) {
    error.value = cause instanceof ApiError ? t(errorMessageKey(cause.code)) : t('auth.verificationLinkInvalid')
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <main class="flex min-h-svh items-center justify-center bg-muted/40 p-6">
    <Card class="w-full max-w-md">
      <CardHeader><CardTitle>{{ t('auth.verifyEmailTitle') }}</CardTitle><CardDescription>{{ t('auth.verifyEmailDescription') }}</CardDescription></CardHeader>
      <CardContent class="flex flex-col gap-4 text-sm">
        <p v-if="loading">{{ t('resource.loading') }}</p>
        <p v-else-if="pending">{{ t('auth.verificationSent') }}</p>
        <p v-else-if="verified" class="text-green-600">{{ t('auth.verificationSuccess') }}</p>
        <p v-else class="text-destructive" role="alert">{{ error }}</p>
        <div class="flex gap-2"><Button @click="router.push({ name: 'login' })">{{ t('auth.login') }}</Button><Button variant="outline" @click="router.push({ path: '/' })">{{ t('member.nav.home') }}</Button></div>
      </CardContent>
    </Card>
  </main>
</template>
