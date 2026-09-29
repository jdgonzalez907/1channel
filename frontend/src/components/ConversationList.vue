<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { asApiError, listConversations, type ApiError } from '@/api/client'
import type { ConversationListItem, ConversationTab } from '@/api/types'
import { formatDateTime } from '@/format'

const props = defineProps<{
  agentId: string
  reloadKey: number
  selectedId: string | null
}>()

const emit = defineEmits<{
  select: [item: ConversationListItem]
  'open-simulator': []
}>()

const status = ref<ConversationTab>('open')
const contactInput = ref('')
const contactFilter = ref('')

const items = ref<ConversationListItem[]>([])
const nextSentAt = ref<string | null>(null)
const nextId = ref<string | null>(null)
const loading = ref(false)
const loadingMore = ref(false)
const error = ref<ApiError | null>(null)

const hasMore = computed(() => nextSentAt.value !== null && nextId.value !== null)

async function load(reset: boolean) {
  error.value = null
  loading.value = reset
  loadingMore.value = !reset
  try {
    const response = await listConversations(props.agentId, status.value, {
      externalContactId: contactFilter.value || undefined,
      before_sent_at: reset ? undefined : (nextSentAt.value ?? undefined),
      before_id: reset ? undefined : (nextId.value ?? undefined),
    })
    items.value = reset ? response.items : items.value.concat(response.items)
    nextSentAt.value = response.next_before_sent_at
    nextId.value = response.next_before_id
  } catch (loadError) {
    error.value = asApiError(loadError)
    if (reset) items.value = []
  } finally {
    loading.value = false
    loadingMore.value = false
  }
}

function changeStatus(next: ConversationTab) {
  status.value = next
  contactInput.value = ''
  contactFilter.value = ''
  void load(true)
}

function search() {
  contactFilter.value = contactInput.value.trim()
  void load(true)
}

function clear() {
  contactInput.value = ''
  contactFilter.value = ''
  void load(true)
}

watch(
  () => [props.agentId, props.reloadKey],
  () => void load(true),
  { immediate: true },
)
</script>

<template>
  <div class="list">
    <div class="controls">
      <div class="statuses">
        <label><input type="radio" :checked="status === 'open'" @change="changeStatus('open')" /> open</label>
        <label><input type="radio" :checked="status === 'finished'" @change="changeStatus('finished')" /> finished</label>
      </div>

      <form class="contact-filter" @submit.prevent="search">
        <input v-model="contactInput" type="text" placeholder="external contact id" />
        <button type="submit">Buscar</button>
        <button type="button" @click="clear">Limpiar</button>
      </form>

      <button type="button" @click="emit('open-simulator')">Simular webhook</button>
    </div>

    <div class="body">
      <p v-if="loading">Cargando...</p>
      <p v-else-if="error" class="error" role="alert">
        {{ error.title }}<span v-if="error.detail"> — {{ error.detail }}</span>
      </p>
      <p v-else-if="items.length === 0">No hay conversaciones.</p>

      <ul v-else class="items">
        <li v-for="item in items" :key="item.id">
          <button
            type="button"
            class="item"
            :class="{ selected: item.id === selectedId }"
            @click="emit('select', item)"
          >
            <span class="row">
              <strong class="contact">{{ item.contact.label }}</strong>
              <span class="status" :class="`status-${item.status}`">{{ item.status }}</span>
            </span>
            <span class="preview">
              {{ item.last_message.owner === 'agent' ? 'Agente' : 'Contacto' }}:
              {{ item.last_message.text === null ? '(mensaje eliminado)' : item.last_message.text }}
            </span>
            <span class="meta">
              <span>{{ formatDateTime(item.last_message.sent_at) }}</span>
              <span
                v-if="item.unread_count > 0"
                class="unread"
                :class="{
                  'unread-open': item.status !== 'expired',
                  'unread-muted': item.status === 'expired',
                }"
                >{{ item.unread_count }}</span
              >
            </span>
          </button>
        </li>
      </ul>
    </div>

    <div v-if="hasMore && !loading && !error" class="more">
      <button type="button" :disabled="loadingMore" @click="load(false)">
        {{ loadingMore ? 'Cargando...' : 'Cargar más' }}
      </button>
    </div>
  </div>
</template>

<style scoped>
.list {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
}

.controls {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  padding: 0.75rem;
  border-bottom: 1px solid #d1d5db;
}

.statuses,
.contact-filter {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.body {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 0.5rem 0.75rem;
}

.items {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  width: 100%;
  text-align: left;
  padding: 0.5rem;
  border: 1px solid #d1d5db;
  border-radius: 0.35rem;
  background: transparent;
  cursor: pointer;
  font: inherit;
}

.item:hover {
  background: #f9fafb;
}

.item.selected {
  border-color: #2563eb;
  box-shadow: inset 0 0 0 1px #2563eb;
}

.row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
}

.contact {
  font-weight: 700;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.preview {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: #4b5563;
}

.meta {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  font-size: 0.8rem;
  color: #6b7280;
}

.status {
  flex-shrink: 0;
  padding: 0.05rem 0.45rem;
  border-radius: 0.65rem;
  font-size: 0.72rem;
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
  margin-left: auto;
  min-width: 1.15rem;
  text-align: center;
  padding: 0.05rem 0.4rem;
  border-radius: 0.65rem;
  background: #dc2626;
  color: #ffffff;
  font-weight: 700;
}

.unread-muted {
  background: #9ca3af;
}

.unread-open {
  background: #16a34a;
}

.more {
  padding: 0.5rem 0.75rem;
  border-top: 1px solid #d1d5db;
}

.error {
  margin: 0;
  color: #b91c1c;
}
</style>
