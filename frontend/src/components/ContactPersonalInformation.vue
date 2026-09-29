<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import {
  asApiError,
  getContact,
  getPersonalInformationByDocument,
  saveContactPersonalInformation,
  type ApiError,
} from '@/api/client'
import type { PersonalInformation, SavePersonalInformationInput } from '@/api/types'

const props = defineProps<{
  agentId: string
  contactId: string | null
}>()

const emit = defineEmits<{ saved: [] }>()

const MAX_IDENTIFICATION = 100
const MAX_NAME = 100
const MAX_PHONE = 100
const MAX_EMAIL = 254
const MAX_ADDRESS = 254

const identificationNumber = ref('')
const firstName = ref('')
const lastName = ref('')
const phoneNumber = ref('')
const email = ref('')
const address = ref('')

const loaded = ref<PersonalInformation | null>(null)
const loading = ref(false)
const searching = ref(false)
const saving = ref(false)
const error = ref<ApiError | null>(null)
const message = ref('')
let requestToken = 0

const segmenter =
  typeof Intl !== 'undefined' && 'Segmenter' in Intl ? new Intl.Segmenter() : null

function graphemeLength(value: string): number {
  if (segmenter) return [...segmenter.segment(value)].length
  return [...value].length
}

function optionalLengthError(value: string, max: number): string | null {
  const trimmed = value.trim()
  if (trimmed.length === 0) return null
  if (graphemeLength(trimmed) > max) return `Máximo ${max}`
  return null
}

function identificationError(): string | null {
  const trimmed = identificationNumber.value.trim()
  if (trimmed.length === 0) return 'Requerido'
  if (graphemeLength(trimmed) > MAX_IDENTIFICATION) return `Máximo ${MAX_IDENTIFICATION}`
  if (!/^[A-Za-z0-9]+$/.test(trimmed)) return 'Solo letras y números'
  return null
}

const fieldErrors = computed<Record<string, string | null>>(() => ({
  identification: identificationError(),
  firstName: optionalLengthError(firstName.value, MAX_NAME),
  lastName: optionalLengthError(lastName.value, MAX_NAME),
  phoneNumber: optionalLengthError(phoneNumber.value, MAX_PHONE),
  email: optionalLengthError(email.value, MAX_EMAIL),
  address: optionalLengthError(address.value, MAX_ADDRESS),
}))

const hasErrors = computed(() => Object.values(fieldErrors.value).some((value) => value !== null))
const canSearch = computed(
  () => !searching.value && fieldErrors.value.identification === null,
)

function applyPersonalInformation(personalInformation: PersonalInformation) {
  loaded.value = personalInformation
  identificationNumber.value = personalInformation.identification_number
  firstName.value = personalInformation.first_name ?? ''
  lastName.value = personalInformation.last_name ?? ''
  phoneNumber.value = personalInformation.phone_number ?? ''
  email.value = personalInformation.email ?? ''
  address.value = personalInformation.address ?? ''
}

function clearDataFields() {
  loaded.value = null
  firstName.value = ''
  lastName.value = ''
  phoneNumber.value = ''
  email.value = ''
  address.value = ''
}

function reset() {
  loaded.value = null
  identificationNumber.value = ''
  clearDataFields()
  error.value = null
  message.value = ''
}

async function load(contactId: string) {
  const token = ++requestToken
  loading.value = true
  error.value = null
  message.value = ''
  try {
    const contact = await getContact(props.agentId, contactId)
    if (token !== requestToken) return
    if (contact.personal_information) {
      applyPersonalInformation(contact.personal_information)
    } else {
      clearDataFields()
      identificationNumber.value = ''
    }
  } catch (loadError) {
    if (token !== requestToken) return
    reset()
    error.value = asApiError(loadError)
  } finally {
    if (token === requestToken) loading.value = false
  }
}

async function search() {
  if (!canSearch.value) return
  searching.value = true
  error.value = null
  message.value = ''
  try {
    const personalInformation = await getPersonalInformationByDocument(props.agentId, identificationNumber.value.trim())
    applyPersonalInformation(personalInformation)
    message.value = 'Persona encontrada. Guardá para asociarla al contacto.'
  } catch (searchError) {
    const apiError = asApiError(searchError)
    if (apiError.status === 404) {
      clearDataFields()
      message.value = 'No existe una persona con ese documento. Completá los datos y guardá.'
    } else {
      error.value = apiError
    }
  } finally {
    searching.value = false
  }
}

async function save() {
  const contactId = props.contactId
  if (!contactId || hasErrors.value) return
  saving.value = true
  error.value = null
  message.value = ''
  const input: SavePersonalInformationInput = {
    identification_number: identificationNumber.value.trim(),
    first_name: firstName.value.trim(),
    last_name: lastName.value.trim(),
    phone_number: phoneNumber.value.trim(),
    email: email.value.trim(),
    address: address.value.trim(),
  }
  try {
    const personalInformation = await saveContactPersonalInformation(props.agentId, contactId, input)
    applyPersonalInformation(personalInformation)
    message.value = 'Persona guardada y asociada al contacto.'
    emit('saved')
  } catch (saveError) {
    error.value = asApiError(saveError)
  } finally {
    saving.value = false
  }
}

watch(
  () => [props.agentId, props.contactId],
  () => {
    const contactId = props.contactId
    if (!contactId) {
      reset()
      return
    }
    void load(contactId)
  },
  { immediate: true },
)
</script>

<template>
  <div class="person">
    <header class="head">
      <strong>Información personal</strong>
      <span v-if="loaded" class="badge">asociada</span>
    </header>

    <p v-if="!contactId" class="placeholder">Seleccioná una conversación.</p>
    <p v-else-if="loading" class="placeholder">Cargando...</p>

    <form v-else class="form" @submit.prevent="save">
      <label class="field">
        <span>identification_number</span>
        <span class="row">
          <input v-model="identificationNumber" type="text" placeholder="solo letras y números" />
          <button type="button" :disabled="!canSearch" @click="search">
            {{ searching ? 'Buscando...' : 'Buscar' }}
          </button>
        </span>
        <small v-if="fieldErrors.identification" class="error-text">{{ fieldErrors.identification }}</small>
      </label>

      <label class="field">
        <span>first_name</span>
        <input v-model="firstName" type="text" />
        <small v-if="fieldErrors.firstName" class="error-text">{{ fieldErrors.firstName }}</small>
      </label>

      <label class="field">
        <span>last_name</span>
        <input v-model="lastName" type="text" />
        <small v-if="fieldErrors.lastName" class="error-text">{{ fieldErrors.lastName }}</small>
      </label>

      <label class="field">
        <span>phone_number</span>
        <input v-model="phoneNumber" type="text" />
        <small v-if="fieldErrors.phoneNumber" class="error-text">{{ fieldErrors.phoneNumber }}</small>
      </label>

      <label class="field">
        <span>email</span>
        <input v-model="email" type="text" />
        <small v-if="fieldErrors.email" class="error-text">{{ fieldErrors.email }}</small>
      </label>

      <label class="field">
        <span>address</span>
        <input v-model="address" type="text" />
        <small v-if="fieldErrors.address" class="error-text">{{ fieldErrors.address }}</small>
      </label>

      <p v-if="message" class="message">{{ message }}</p>
      <p v-if="error" class="error-text" role="alert">
        {{ error.title }}<span v-if="error.detail"> — {{ error.detail }}</span>
      </p>

      <button type="submit" :disabled="saving || hasErrors">
        {{ saving ? 'Guardando...' : 'Guardar' }}
      </button>
    </form>
  </div>
</template>

<style scoped>
.person {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
  overflow-y: auto;
}

.head {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-shrink: 0;
  padding: 0.75rem;
  border-bottom: 1px solid #d1d5db;
}

.badge {
  padding: 0.05rem 0.45rem;
  border-radius: 0.65rem;
  font-size: 0.72rem;
  font-weight: 600;
  background: #dcfce7;
  color: #166534;
}

.placeholder {
  padding: 0.75rem;
  margin: 0;
}

.form {
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
  padding: 0.75rem;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}

.field > span:first-child {
  font-size: 0.78rem;
  color: #4b5563;
}

.field .row {
  display: flex;
  gap: 0.4rem;
}

.field .row input {
  flex: 1;
  min-width: 0;
}

.field input {
  font: inherit;
  min-width: 0;
}

.error-text {
  margin: 0;
  color: #b91c1c;
  font-size: 0.78rem;
}

.message {
  margin: 0;
  color: #166534;
  font-size: 0.82rem;
}
</style>
