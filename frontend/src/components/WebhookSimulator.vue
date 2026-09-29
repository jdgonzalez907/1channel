<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { asApiError, simulateWebhook, type ApiError } from '@/api/client'
import type { WebhookEvent } from '@/api/types'

const props = defineProps<{
  open: boolean
  contactExternalId: string
}>()

const emit = defineEmits<{
  close: []
  sent: [externalContactId: string]
}>()

const event = ref<WebhookEvent>('message.received')
const contactId = ref('')
const messageId = ref('')
const displayName = ref('')
const text = ref('')
const error = ref<ApiError | null>(null)
const sending = ref(false)

const needsText = computed(
  () => event.value === 'message.received' || event.value === 'message.edited',
)

const setsDisplayName = computed(() => event.value === 'message.received')

watch(
  () => props.open,
  (isOpen) => {
    if (!isOpen) return
    event.value = 'message.received'
    contactId.value = props.contactExternalId
    messageId.value = ''
    displayName.value = ''
    text.value = ''
    error.value = null
  },
)

async function submit() {
  error.value = null
  sending.value = true
  try {
    await simulateWebhook({
      event: event.value,
      external_contact_id: contactId.value,
      external_message_id: messageId.value,
      ...(setsDisplayName.value && displayName.value.trim() !== ''
        ? { display_name: displayName.value.trim() }
        : {}),
      ...(needsText.value ? { text: text.value } : {}),
    })
    emit('sent', contactId.value)
  } catch (sendError) {
    error.value = asApiError(sendError)
  } finally {
    sending.value = false
  }
}
</script>

<template>
  <div v-if="open" class="overlay" @click.self="emit('close')">
    <form class="modal" @submit.prevent="submit">
      <h2>Simular evento del contacto</h2>

      <label>
        Evento
        <select v-model="event">
          <option value="message.received">message.received</option>
          <option value="message.edited">message.edited</option>
          <option value="message.deleted">message.deleted</option>
          <option value="message.read">message.read</option>
        </select>
      </label>

      <label>
        external_contact_id
        <input v-model="contactId" type="text" />
      </label>

      <label v-if="setsDisplayName">
        display_name (opcional)
        <input v-model="displayName" type="text" />
      </label>

      <label>
        external_message_id
        <input v-model="messageId" type="text" />
      </label>
      <p class="hint">
        Para edit/deleted/read hay que pegar el id externo a mano: la API de lectura no lo expone.
      </p>

      <label v-if="needsText">
        text
        <textarea v-model="text" rows="3"></textarea>
      </label>

      <p v-if="error" class="error" role="alert">
        {{ error.title }}<span v-if="error.detail"> — {{ error.detail }}</span>
      </p>

      <div class="actions">
        <button type="submit" :disabled="sending">{{ sending ? 'Enviando...' : 'Enviar' }}</button>
        <button type="button" @click="emit('close')">Cerrar</button>
      </div>
    </form>
  </div>
</template>

<style scoped>
.overlay {
  position: fixed;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.35);
}

.modal {
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
  width: 24rem;
  max-width: calc(100vw - 2rem);
  padding: 1rem;
  background: #ffffff;
  color: #111827;
  border: 1px solid #d1d5db;
  border-radius: 0.5rem;
}

.modal label {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.modal input,
.modal select,
.modal textarea {
  font: inherit;
}

.hint {
  margin: 0;
  font-size: 0.8rem;
}

.actions {
  display: flex;
  gap: 0.5rem;
}

.error {
  margin: 0;
}
</style>
