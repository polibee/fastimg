<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { Languages } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { ApiError, errorMessageKey } from '@/lib/api'
import { resolveLoginRedirect } from '@/lib/login-redirect'
import { useAuthStore } from '@/stores/auth'
import { generatedApi, type RegistrationPolicy } from '@/generated/api'
import TurnstileWidget from '../components/TurnstileWidget.vue'

const { t, locale } = useI18n()
const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const email = ref('')
const password = ref('')
const errorMessage = ref('')
const isSubmitting = ref(false)
const policy = ref<RegistrationPolicy>()
const turnstileToken = ref('')

onMounted(async () => {
  try { policy.value = await generatedApi.registrationPolicy() } catch { /* login remains usable when policy discovery is unavailable */ }
})

function toggleLocale() {
  locale.value = locale.value === 'zh-CN' ? 'en-US' : 'zh-CN'
  localStorage.setItem('locale', locale.value)
}

async function submit() {
  errorMessage.value = ''
  isSubmitting.value = true
  try {
    await auth.login(email.value, password.value, turnstileToken.value)
    await router.replace(resolveLoginRedirect(route.query.redirect))
  } catch (error) {
    errorMessage.value = error instanceof ApiError ? t(errorMessageKey(error.code)) : t('errors.unknown')
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <main class="relative flex min-h-svh items-center justify-center bg-muted/40 p-6">
    <Button class="absolute right-6 top-6" variant="ghost" size="sm" :aria-label="t('auth.language')" @click="toggleLocale">
      <Languages data-icon="inline-start" />
      {{ locale === 'zh-CN' ? 'EN' : '中文' }}
    </Button>
    <Card class="w-full max-w-sm">
      <CardHeader>
        <CardTitle>{{ t('auth.loginTitle') }}</CardTitle>
        <CardDescription>{{ t('auth.loginDescription') }}</CardDescription>
      </CardHeader>
      <CardContent>
        <form class="flex flex-col gap-4" @submit.prevent="submit">
          <div class="flex flex-col gap-2">
            <Label for="email">{{ t('auth.email') }}</Label>
            <Input id="email" v-model="email" type="email" autocomplete="username" required />
          </div>
          <div class="flex flex-col gap-2">
            <Label for="password">{{ t('auth.password') }}</Label>
            <Input id="password" v-model="password" type="password" autocomplete="current-password" required />
          </div>
          <TurnstileWidget v-if="policy?.login_turnstile" :site-key="policy.turnstile_site_key" v-model:token="turnstileToken" />
          <p v-if="errorMessage" class="text-sm text-destructive" role="alert">{{ errorMessage }}</p>
          <Button type="submit" :disabled="isSubmitting">
            {{ isSubmitting ? t('auth.loggingIn') : t('auth.login') }}
          </Button>
          <Button type="button" variant="ghost" @click="router.push({ name: 'register' })">{{ t('auth.createAccount') }}</Button>
        </form>
      </CardContent>
    </Card>
  </main>
</template>
