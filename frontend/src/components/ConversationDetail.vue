<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import {
  asApiError,
  getConversation,
  markRead,
  resolveConversation,
  sendMessage,
  type ApiError,
} from '@/api/client'
import type { ConversationDetail } from '@/api/types'
import { formatDateTime } from '@/format'

const props = defineProps<{
  agentId: string
  conversationId: string | null
  reloadKey: number
}>()

const emit = defineEmits<{ changed: [] }>()

const detail = ref<ConversationDetail | null>(null)
const loading = ref(false)
const loadingMore = ref(false)
const error = ref<ApiError | null>(null)
const actionError = ref<ApiError | null>(null)
const replyText = ref('')
const messagesEl = ref<HTMLElement | null>(null)
let requestToken = 0

async function scrollToBottom() {
  await nextTick()
  const container = messagesEl.value
  if (container) container.scrollTop = container.scrollHeight
}

const isOpen = computed(
  () => detail.value?.status === 'pending' || detail.value?.status === 'assigned',
)
const canResolve = computed(() => detail.value?.status === 'assigned')
const hasOlder = computed(
  () => detail.value?.next_before_sent_at !== null && detail.value?.next_before_id !== null,
)

async function load() {
  const conversationId = props.conversationId
  if (!conversationId) {
    detail.value = null
    error.value = null
    return
  }

  const token = ++requestToken
  loading.value = true
  error.value = null
  actionError.value = null
  replyText.value = ''

  try {
    const loaded = await getConversation(props.agentId, conversationId)
    if (token !== requestToken) return
    detail.value = loaded

    if (loaded.agent_id === props.agentId && loaded.unread_count > 0) {
      try {
        await markRead(props.agentId, loaded.id)
        if (token !== requestToken) return
        const readAt = new Date().toISOString()
        detail.value = {
          ...loaded,
          unread_count: 0,
          messages: loaded.messages.map((message) =>
            message.owner === 'contact' && message.read_at === null
              ? { ...message, read_at: readAt }
              : message,
          ),
        }
        emit('changed')
      } catch (markError) {
        actionError.value = asApiError(markError)
      }
    }
  } catch (loadError) {
    if (token !== requestToken) return
    detail.value = null
    error.value = asApiError(loadError)
  } finally {
    if (token === requestToken) loading.value = false
  }

  if (token === requestToken) await scrollToBottom()
}

async function loadOlder() {
  const current = detail.value
  if (!current || current.next_before_sent_at === null || current.next_before_id === null) return
  const container = messagesEl.value
  const previousHeight = container?.scrollHeight ?? 0
  const previousTop = container?.scrollTop ?? 0
  loadingMore.value = true
  actionError.value = null
  try {
    const older = await getConversation(props.agentId, current.id, {
      before_sent_at: current.next_before_sent_at,
      before_id: current.next_before_id,
    })
    detail.value = {
      ...current,
      messages: older.messages.concat(current.messages),
      next_before_sent_at: older.next_before_sent_at,
      next_before_id: older.next_before_id,
    }
    await nextTick()
    const updated = messagesEl.value
    if (updated) updated.scrollTop = updated.scrollHeight - previousHeight + previousTop
  } catch (loadError) {
    actionError.value = asApiError(loadError)
  } finally {
    loadingMore.value = false
  }
}

async function reply() {
  if (!detail.value) return
  actionError.value = null
  try {
    await sendMessage(props.agentId, detail.value.id, replyText.value)
    emit('changed')
    await load()
  } catch (sendError) {
    actionError.value = asApiError(sendError)
  }
}

async function resolve() {
  if (!detail.value) return
  actionError.value = null
  try {
    await resolveConversation(props.agentId, detail.value.id)
    emit('changed')
    await load()
  } catch (resolveError) {
    actionError.value = asApiError(resolveError)
  }
}

watch(
  () => [props.agentId, props.conversationId, props.reloadKey],
  () => void load(),
  { immediate: true },
)
</script>

<template>
  <div class="detail">
    <p v-if="!conversationId" class="placeholder">Seleccioná una conversación.</p>
    <p v-else-if="loading" class="placeholder">Cargando...</p>
    <p v-else-if="error" class="placeholder error" role="alert">
      {{ error.title }}<span v-if="error.detail"> — {{ error.detail }}</span>
    </p>

    <template v-else-if="detail">
      <header class="head">
        <div class="head-info">
          <div class="contact-block">
            <strong>{{ detail.contact.external_id }}</strong>
            <span class="conversation-id">{{ detail.id }}</span>
          </div>
          <span class="status" :class="`status-${detail.status}`">{{ detail.status }}</span>
          <span
            v-if="detail.unread_count > 0"
            class="unread"
            :class="{
              'unread-open': detail.agent_id === agentId,
              'unread-muted': detail.agent_id !== agentId,
            }"
            >{{ detail.unread_count }} sin leer</span
          >
        </div>
        <button v-if="canResolve" type="button" @click="resolve">Resolver</button>
      </header>

      <p v-if="actionError" class="action-error" role="alert">
        {{ actionError.title }}<span v-if="actionError.detail"> — {{ actionError.detail }}</span>
      </p>

      <div ref="messagesEl" class="messages">
        <button v-if="hasOlder" type="button" :disabled="loadingMore" @click="loadOlder">
          {{ loadingMore ? 'Cargando...' : 'Cargar más' }}
        </button>

        <ul>
          <li
            v-for="message in detail.messages"
            :key="message.id"
            class="message"
            :class="message.owner"
          >
            <span class="owner">{{ message.owner === 'agent' ? 'Agente' : 'Contacto' }}</span>
            <span class="text">
              {{ message.status === 'deleted' || message.text === null ? '(mensaje eliminado)' : message.text }}
            </span>
            <span class="date">
              {{ formatDateTime(message.sent_at) }}
              <span v-if="message.read_at && message.status !== 'deleted'" class="read">leído</span>
              <span v-if="message.edited_at && message.status !== 'deleted'" class="edited">editado</span>
            </span>
          </li>
        </ul>
      </div>

      <form v-if="isOpen" class="reply" @submit.prevent="reply">
        <textarea v-model="replyText" rows="3" placeholder="Escribí una respuesta"></textarea>
        <button type="submit">Enviar</button>
      </form>
    </template>
  </div>
</template>

<style scoped>
.detail {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
}

.placeholder {
  padding: 0.75rem;
  margin: 0;
}

.head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  flex-shrink: 0;
  padding: 0.75rem;
  border-bottom: 1px solid #d1d5db;
}

.head-info {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex-wrap: wrap;
}

.conversation-id {
  font-family: ui-monospace, monospace;
  font-size: 0.78rem;
  color: #6b7280;
}

.contact-block {
  display: flex;
  flex-direction: column;
}

.messages {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 0.75rem;
}

.messages ul {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.message {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
  max-width: 78%;
  padding: 0.45rem 0.6rem;
  border-radius: 0.6rem;
  overflow-wrap: anywhere;
}

.message.contact {
  align-self: flex-start;
  background: #f3f4f6;
}

.message.agent {
  align-self: flex-end;
  background: #dbeafe;
  text-align: right;
}

.owner {
  font-size: 0.75rem;
  color: #6b7280;
}

.text {
  white-space: pre-wrap;
}

.date {
  font-size: 0.72rem;
  color: #6b7280;
}

.read,
.edited {
  font-weight: 600;
}

.read {
  color: #16a34a;
}

.edited {
  color: #d97706;
}

.reply {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  align-items: flex-start;
  flex-shrink: 0;
  padding: 0.75rem;
  border-top: 1px solid #d1d5db;
}

.reply textarea {
  width: 100%;
  box-sizing: border-box;
  font: inherit;
}

.status {
  padding: 0.05rem 0.45rem;
  border-radius: 0.65rem;
  font-size: 0.75rem;
  font-weight: 600;
}

.status-pending {
  background: #fef3c7;
  color: #92400e;
}

.status-assigned {
  background: #dbeafe;
  color: #1e40af;
}

.status-resolved {
  background: #dcfce7;
  color: #166534;
}

.status-expired {
  background: #f3f4f6;
  color: #6b7280;
}

.unread {
  padding: 0.05rem 0.4rem;
  border-radius: 0.65rem;
  background: #dc2626;
  color: #ffffff;
  font-size: 0.75rem;
  font-weight: 700;
}

.unread-muted {
  background: #9ca3af;
}

.unread-open {
  background: #16a34a;
}

.action-error {
  margin: 0;
  padding: 0.5rem 0.75rem;
  color: #b91c1c;
}

.error {
  color: #b91c1c;
}
</style>
