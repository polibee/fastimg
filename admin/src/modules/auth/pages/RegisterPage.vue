<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { ApiError, errorMessageKey } from '@/lib/api'
import { generatedApi, type RegistrationPolicy } from '@/generated/api'
import TurnstileWidget from '../components/TurnstileWidget.vue'

const { t, locale } = useI18n()
const router = useRouter()
const name = ref('')
const email = ref('')
const password = ref('')
const passwordConfirmation = ref('')
const turnstileToken = ref('')
const policy = ref<RegistrationPolicy>()
const loading = ref(true)
const submitting = ref(false)
const errorMessage = ref('')

function toggleLocale() {
  locale.value = locale.value === 'zh-CN' ? 'en-US' : 'zh-CN'
  localStorage.setItem('locale', locale.value)
}

onMounted(async () => {
  try {
    policy.value = await generatedApi.registrationPolicy()
  } catch (error) {
    errorMessage.value = error instanceof ApiError ? t(errorMessageKey(error.code)) : t('errors.unknown')
  } finally {
    loading.value = false
  }
})

async function submit() {
  errorMessage.value = ''
  if (password.value !== passwordConfirmation.value) {
    errorMessage.value = t('auth.passwordMismatch')
    return
  }
  submitting.value = true
  try {
    const result = await generatedApi.register({ name: name.value, email: email.value, password: password.value, password_confirmation: passwordConfirmation.value, turnstile_token: turnstileToken.value })
    if (result.verification_required) {
      await router.replace({ name: 'verify-email', query: { pending: '1' } })
    } else {
      await router.replace({ name: 'login', query: { registered: '1' } })
    }
  } catch (error) {
    errorMessage.value = error instanceof ApiError ? t(errorMessageKey(error.code)) : t('errors.unknown')
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <main class="relative flex min-h-svh items-center justify-center bg-muted/40 p-6">
    <Button class="absolute right-6 top-6" variant="ghost" size="sm" :aria-label="t('auth.language')" @click="toggleLocale">{{ locale === 'zh-CN' ? 'EN' : '中文' }}</Button>
    <Card class="w-full max-w-md">
      <CardHeader>
        <CardTitle>{{ t('auth.registerTitle') }}</CardTitle>
        <CardDescription>{{ t('auth.registerDescription') }}</CardDescription>
      </CardHeader>
      <CardContent>
        <div v-if="loading" class="text-sm text-muted-foreground">{{ t('resource.loading') }}</div>
        <div v-else-if="policy && !policy.registration_enabled" class="flex flex-col gap-4 text-sm">
          <p>{{ t('auth.registrationDisabled') }}</p>
          <Button variant="outline" @click="router.push({ name: 'login' })">{{ t('auth.login') }}</Button>
        </div>
        <form v-else class="flex flex-col gap-4" @submit.prevent="submit">
          <div class="flex flex-col gap-2"><Label for="register-name">{{ t('auth.name') }}</Label><Input id="register-name" v-model="name" autocomplete="name" required /></div>
          <div class="flex flex-col gap-2"><Label for="register-email">{{ t('auth.email') }}</Label><Input id="register-email" v-model="email" type="email" autocomplete="email" required /></div>
          <div class="flex flex-col gap-2"><Label for="register-password">{{ t('auth.password') }}</Label><Input id="register-password" name="password" v-model="password" type="password" autocomplete="new-password" minlength="8" required /></div>
          <div class="flex flex-col gap-2"><Label for="register-password-confirmation">{{ t('auth.passwordConfirmation') }}</Label><Input id="register-password-confirmation" name="password_confirmation" v-model="passwordConfirmation" type="password" autocomplete="new-password" minlength="8" required /></div>
          <TurnstileWidget v-if="policy?.registration_turnstile" :site-key="policy.turnstile_site_key" v-model:token="turnstileToken" />
          <p v-if="policy?.email_verification_required" class="text-xs text-muted-foreground">{{ t('auth.emailVerificationNotice') }}</p>
          <p v-if="errorMessage" class="text-sm text-destructive" role="alert">{{ errorMessage }}</p>
          <Button type="submit" :disabled="submitting">{{ submitting ? t('auth.registering') : t('auth.register') }}</Button>
          <Button type="button" variant="ghost" @click="router.push({ name: 'login' })">{{ t('auth.alreadyHaveAccount') }}</Button>
        </form>
      </CardContent>
    </Card>
  </main>
</template>
