<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft, Copy, Dices, Eye, EyeOff, Save } from '@lucide/vue'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Field, FieldGroup, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupInput } from '@/components/ui/input-group'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { ApiError, errorMessageKey } from '@/lib/api'
import { generatedApi, type AdminPlanSummary } from '@/generated/api'
import { createResourceForm, serializeResourceForm, type ResourceFormField } from '@/lib/resource-form'
import { generatePassword } from '@/lib/password-generator'
import { userStatusLabelKey, type UserStatus } from '@/lib/user-status'
import { useAuthStore } from '@/stores/auth'
import { useI18n } from 'vue-i18n'

interface ResourceManifest { name: string; label: string; fields: ResourceFormField[] }

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const editing = computed(() => Boolean(route.params.id))
const loading = ref(editing.value)
const saving = ref(false)
const error = ref('')
const fields = ref<ResourceFormField[]>([])
const form = ref<Record<string, any>>({})
const showPassword = ref(false)
const copiedPassword = ref(false)
const subscriptionLoading = ref(false)
const subscriptionSaving = ref(false)
const subscriptionError = ref('')
const subscriptionPlans = ref<AdminPlanSummary[]>([])
const selectedPlanID = ref('')
const visibleFields = computed(() => fields.value.filter((field) => field.name !== 'locale'))

function localizedError(value: unknown) { return value instanceof ApiError ? t(errorMessageKey(value.code)) : t('errors.unknown') }
function fieldId(field: ResourceFormField) { return `resource-field-${field.name}` }
function isPassword(field: ResourceFormField) { return field.type === 'password' }
function isBoolean(field: ResourceFormField) { return field.type === 'boolean' }
function isSelect(field: ResourceFormField) { return field.type === 'select' }
function isNumber(field: ResourceFormField) { return field.type === 'number' }
function inputType(field: ResourceFormField) { return field.type === 'email' || field.type === 'password' || field.type === 'number' || field.type === 'date' ? field.type : 'text' }
function fieldRequired(field: ResourceFormField) { return !editing.value && field.name !== 'locale' }
function fieldDescription(field: ResourceFormField) { return isPassword(field) ? (editing.value ? t('resource.passwordHint') : t('resource.passwordRequired')) : '' }
function statusOptionLabel(value: string) { return t(userStatusLabelKey(value as UserStatus)) }
function generateUserPassword() {
  form.value.password = generatePassword()
  showPassword.value = true
  copiedPassword.value = false
}
async function copyUserPassword() {
  if (!form.value.password) return
  await navigator.clipboard.writeText(String(form.value.password))
  copiedPassword.value = true
}

onMounted(async () => {
  if (!auth.token) return
  try {
    const manifests = await generatedApi.resourceRegistry(auth.token)
    const manifest = manifests.find((item) => item.name === 'users')
    if (!manifest) throw new Error('users manifest missing')
    fields.value = manifest.fields
    const record = editing.value ? await generatedApi.resourceShow<Record<string, unknown>>('users', String(route.params.id), auth.token) : {}
    form.value = createResourceForm(fields.value, record)
    if (editing.value) {
      subscriptionLoading.value = true
      const subscription = await generatedApi.adminSubscription(String(route.params.id), auth.token)
      subscriptionPlans.value = subscription.plans
      selectedPlanID.value = String(subscription.subscription.plan_id)
    }
  } catch (value) {
    error.value = localizedError(value)
  } finally {
    subscriptionLoading.value = false
    loading.value = false
  }
})

async function saveSubscription() {
  if (!auth.token || !editing.value || !selectedPlanID.value) return
  subscriptionSaving.value = true
  subscriptionError.value = ''
  try {
    const response = await generatedApi.updateAdminSubscription(String(route.params.id), Number(selectedPlanID.value), auth.token)
    selectedPlanID.value = String(response.subscription.plan_id)
  } catch (value) {
    subscriptionError.value = localizedError(value)
  } finally {
    subscriptionSaving.value = false
  }
}

async function submit() {
  if (!auth.token) return
  saving.value = true
  error.value = ''
  try {
    const payload = serializeResourceForm(fields.value, form.value)
    if (editing.value) await generatedApi.resourceUpdate('users', String(route.params.id), payload, auth.token)
    else await generatedApi.resourceCreate('users', payload, auth.token)
    await router.push('/users')
  } catch (value) {
    error.value = localizedError(value)
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="flex flex-col gap-6">
    <div class="flex items-center gap-3">
      <Button variant="ghost" size="icon" :aria-label="t('resource.back')" @click="router.push('/users')"><ArrowLeft /></Button>
      <div><h1 class="text-2xl font-semibold tracking-tight">{{ editing ? t('resource.editUser') : t('resource.createUser') }}</h1><p class="text-sm text-muted-foreground">{{ t('resource.userFormDescription') }}</p></div>
    </div>
    <Alert v-if="error" variant="destructive"><AlertTitle>{{ t('states.errorTitle') }}</AlertTitle><AlertDescription>{{ error }}</AlertDescription></Alert>
    <Card v-if="loading"><CardContent class="py-8">{{ t('resource.loading') }}</CardContent></Card>
    <Card v-else>
      <CardHeader><CardTitle>{{ editing ? t('resource.editUser') : t('resource.createUser') }}</CardTitle><CardDescription>{{ t('resource.userFormDescription') }}</CardDescription></CardHeader>
      <CardContent><form class="grid gap-5 sm:max-w-xl" @submit.prevent="submit">
        <FieldGroup>
          <Field v-for="field in visibleFields" :key="field.name">
            <FieldLabel :for="fieldId(field)">{{ field.label }}</FieldLabel>
            <Switch v-if="isBoolean(field)" :id="fieldId(field)" v-model="form[field.name]" />
            <Select v-else-if="isSelect(field)" v-model="form[field.name]">
              <SelectTrigger :id="fieldId(field)"><SelectValue /></SelectTrigger>
              <SelectContent><SelectItem v-for="option in field.options || []" :key="option.value" :value="option.value">{{ statusOptionLabel(option.value) }}</SelectItem></SelectContent>
            </Select>
            <InputGroup v-else-if="isPassword(field)">
              <InputGroupInput :id="fieldId(field)" v-model="form[field.name]" :type="showPassword ? 'text' : 'password'" :required="fieldRequired(field)" autocomplete="new-password" />
              <InputGroupAddon align="inline-end">
                <InputGroupButton :aria-label="showPassword ? t('resource.hidePassword') : t('resource.showPassword')" :title="showPassword ? t('resource.hidePassword') : t('resource.showPassword')" @click="showPassword = !showPassword"><EyeOff v-if="showPassword" /><Eye v-else /></InputGroupButton>
                <InputGroupButton :aria-label="t('resource.generatePassword')" :title="t('resource.generatePassword')" @click="generateUserPassword"><Dices /></InputGroupButton>
                <InputGroupButton :aria-label="copiedPassword ? t('resource.copiedPassword') : t('resource.copyPassword')" :title="copiedPassword ? t('resource.copiedPassword') : t('resource.copyPassword')" @click="copyUserPassword"><Copy /></InputGroupButton>
              </InputGroupAddon>
            </InputGroup>
            <Input v-else :id="fieldId(field)" v-model="form[field.name]" :type="inputType(field)" :required="fieldRequired(field)" :autocomplete="isPassword(field) ? 'new-password' : undefined" :step="isNumber(field) ? '1' : undefined" />
            <p v-if="fieldDescription(field)" class="text-xs text-muted-foreground">{{ fieldDescription(field) }}</p>
          </Field>
        </FieldGroup>
        <div class="flex gap-2"><Button type="submit" :disabled="saving"><Save data-icon="inline-start" />{{ saving ? t('resource.saving') : t('resource.save') }}</Button><Button type="button" variant="outline" @click="router.push('/users')">{{ t('resource.cancel') }}</Button></div>
      </form></CardContent>
    </Card>
    <Card v-if="editing">
      <CardHeader><CardTitle>{{ t('resource.userPlanTitle') }}</CardTitle><CardDescription>{{ t('resource.userPlanDescription') }}</CardDescription></CardHeader>
      <CardContent class="grid gap-4 sm:max-w-xl">
        <Alert v-if="subscriptionError" variant="destructive"><AlertDescription>{{ subscriptionError }}</AlertDescription></Alert>
        <Field>
          <FieldLabel for="user-plan">{{ t('resource.userPlanLabel') }}</FieldLabel>
          <Select v-model="selectedPlanID" :disabled="subscriptionLoading || subscriptionSaving">
            <SelectTrigger id="user-plan"><SelectValue :placeholder="subscriptionLoading ? t('resource.loading') : t('resource.userPlanPlaceholder')" /></SelectTrigger>
            <SelectContent><SelectItem v-for="plan in subscriptionPlans" :key="plan.id" :value="String(plan.id)">{{ plan.name }} · {{ plan.code }}</SelectItem></SelectContent>
          </Select>
          <p class="text-xs text-muted-foreground">{{ t('resource.userPlanHint') }}</p>
        </Field>
        <div><Button type="button" :disabled="subscriptionLoading || subscriptionSaving || !selectedPlanID" @click="saveSubscription">{{ subscriptionSaving ? t('resource.saving') : t('resource.saveUserPlan') }}</Button></div>
      </CardContent>
    </Card>
  </div>
</template>
