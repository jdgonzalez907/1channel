<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { asApiError, getUser, type ApiError } from '@/api/client'
import type { ConversationListItem } from '@/api/types'
import ConversationDetail from '@/components/ConversationDetail.vue'
import ConversationList from '@/components/ConversationList.vue'
import WebhookSimulator from '@/components/WebhookSimulator.vue'

const STORAGE_KEY = 'agent-console.agent-id'

const agentInput = ref('')
const agentId = ref('')
const agentError = ref<ApiError | null>(null)
const connecting = ref(false)

const selected = ref<ConversationListItem | null>(null)
const listReloadKey = ref(0)
const detailReloadKey = ref(0)
const simulatorOpen = ref(false)

onMounted(() => {
  const stored = localStorage.getItem(STORAGE_KEY)
  if (stored) {
    agentInput.value = stored
    void connect(stored)
  }
})

async function connect(id: string) {
  agentError.value = null
  connecting.value = true
  const candidate = id.trim()
  try {
    await getUser(candidate)
    agentId.value = candidate
    localStorage.setItem(STORAGE_KEY, candidate)
  } catch (error) {
    agentId.value = ''
    agentError.value = asApiError(error)
    localStorage.removeItem(STORAGE_KEY)
  } finally {
    connecting.value = false
  }
}

function disconnect() {
  agentId.value = ''
  selected.value = null
  agentError.value = null
}

function onSelect(item: ConversationListItem) {
  selected.value = item
}

function onConversationChanged() {
  listReloadKey.value++
}

function onWebhookSent(externalContactId: string) {
  simulatorOpen.value = false
  listReloadKey.value++
  if (selected.value && selected.value.contact.external_id === externalContactId) {
    detailReloadKey.value++
  }
}
</script>

<template>
  <div class="app">
    <header class="bar">
      <strong>1Channel</strong>
      <form class="agent-form" @submit.prevent="connect(agentInput)">
        <label for="agent-id">Id del agente</label>
        <input id="agent-id" v-model="agentInput" type="text" placeholder="uuid del agente" />
        <button type="submit" :disabled="connecting">
          {{ connecting ? 'Entrando...' : 'Entrar' }}
        </button>
        <button v-if="agentId" type="button" @click="disconnect">Cambiar</button>
      </form>
    </header>

    <p v-if="agentError" class="error" role="alert">
      {{ agentError.title }}<span v-if="agentError.detail"> — {{ agentError.detail }}</span>
    </p>

    <main v-if="agentId" class="panes">
      <section class="pane pane-list">
        <ConversationList
          :agent-id="agentId"
          :reload-key="listReloadKey"
          :selected-id="selected ? selected.id : null"
          @select="onSelect"
          @open-simulator="simulatorOpen = true"
        />
      </section>
      <section class="pane pane-detail">
        <ConversationDetail
          :agent-id="agentId"
          :conversation-id="selected ? selected.id : null"
          :reload-key="detailReloadKey"
          @changed="onConversationChanged"
        />
      </section>
    </main>

    <WebhookSimulator
      v-if="agentId"
      :open="simulatorOpen"
      :contact-external-id="selected ? selected.contact.external_id : ''"
      @close="simulatorOpen = false"
      @sent="onWebhookSent"
    />
  </div>
</template>

<style scoped>
.app {
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow: hidden;
  font-family: system-ui, sans-serif;
}

.bar {
  display: flex;
  align-items: center;
  gap: 1.5rem;
  padding: 0.75rem 1rem;
  border-bottom: 1px solid #d1d5db;
  flex-wrap: wrap;
}

.agent-form {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.error {
  margin: 0;
  padding: 0.5rem 1rem;
  border-bottom: 1px solid #d1d5db;
  color: #b91c1c;
}

.panes {
  display: flex;
  flex: 1;
  min-height: 0;
  overflow: hidden;
}

.pane-list,
.pane-detail {
  display: flex;
  min-height: 0;
  overflow: hidden;
}

.pane-list > *,
.pane-detail > * {
  flex: 1;
  min-width: 0;
  min-height: 0;
}

.pane-list {
  flex: 0 0 22rem;
  border-right: 1px solid #d1d5db;
}

.pane-detail {
  flex: 1;
  min-width: 0;
}

@media (max-width: 720px) {
  .panes {
    flex-direction: column;
  }

  .pane-list {
    flex: 0 0 auto;
    max-height: 45vh;
    border-right: none;
    border-bottom: 1px solid #d1d5db;
  }
}
</style>
