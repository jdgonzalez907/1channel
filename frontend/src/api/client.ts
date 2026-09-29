import type {
  Contact,
  ConversationDetail,
  ConversationListResponse,
  ConversationTab,
  PaginationPosition,
  PersonalInformation,
  ProblemDetails,
  SavePersonalInformationInput,
  User,
  WebhookPayload,
} from './types'

export class ApiError extends Error {
  readonly status: number
  readonly title: string
  readonly detail: string

  constructor(status: number, title: string, detail: string) {
    super(detail || title)
    this.name = 'ApiError'
    this.status = status
    this.title = title
    this.detail = detail
  }
}

export function asApiError(error: unknown): ApiError {
  if (error instanceof ApiError) return error
  return new ApiError(0, 'Error inesperado', error instanceof Error ? error.message : String(error))
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  let response: Response
  try {
    response = await fetch(path, init)
  } catch {
    throw new ApiError(0, 'Error de red', 'No se pudo conectar con la API.')
  }

  if (!response.ok) {
    throw await buildApiError(response)
  }

  if (response.status === 204) {
    return undefined as T
  }

  const body = await response.text()
  return (body ? JSON.parse(body) : undefined) as T
}

async function buildApiError(response: Response): Promise<ApiError> {
  let title = response.statusText || 'Error'
  let detail = ''
  try {
    const problem = (await response.json()) as Partial<ProblemDetails>
    if (typeof problem.title === 'string' && problem.title) title = problem.title
    if (typeof problem.detail === 'string') detail = problem.detail
  } catch {
    // El cuerpo no era problem+json; se usan los valores por defecto.
  }
  return new ApiError(response.status, title, detail)
}

function authHeaders(agentId: string): HeadersInit {
  return { Authorization: `Bearer ${agentId}` }
}

function jsonHeaders(agentId: string): HeadersInit {
  return { ...authHeaders(agentId), 'Content-Type': 'application/json' }
}

function buildQuery(params: Record<string, string | undefined>): string {
  const search = new URLSearchParams()
  for (const [key, value] of Object.entries(params)) {
    if (value !== undefined && value !== '') search.set(key, value)
  }
  const query = search.toString()
  return query ? `?${query}` : ''
}

export function getUser(agentId: string): Promise<User> {
  return request<User>(`/v1/users/${encodeURIComponent(agentId)}`, {
    headers: authHeaders(agentId),
  })
}

export function listConversations(
  agentId: string,
  status: ConversationTab,
  options: { externalContactId?: string } & PaginationPosition = {},
): Promise<ConversationListResponse> {
  const query = buildQuery({
    status,
    external_contact_id: options.externalContactId,
    before_sent_at: options.before_sent_at,
    before_id: options.before_id,
  })
  return request<ConversationListResponse>(`/v1/conversations${query}`, {
    headers: authHeaders(agentId),
  })
}

export function getConversation(
  agentId: string,
  conversationId: string,
  position: PaginationPosition = {},
): Promise<ConversationDetail> {
  const query = buildQuery({
    before_sent_at: position.before_sent_at,
    before_id: position.before_id,
  })
  return request<ConversationDetail>(
    `/v1/conversations/${encodeURIComponent(conversationId)}${query}`,
    { headers: authHeaders(agentId) },
  )
}

export function getContact(agentId: string, contactId: string): Promise<Contact> {
  return request<Contact>(`/v1/contacts/${encodeURIComponent(contactId)}`, {
    headers: authHeaders(agentId),
  })
}

export function getPersonalInformationByDocument(
  agentId: string,
  identificationNumber: string,
): Promise<PersonalInformation> {
  return request<PersonalInformation>(
    `/v1/personal-information/${encodeURIComponent(identificationNumber)}`,
    { headers: authHeaders(agentId) },
  )
}

export function saveContactPersonalInformation(
  agentId: string,
  contactId: string,
  input: SavePersonalInformationInput,
): Promise<PersonalInformation> {
  return request<PersonalInformation>(
    `/v1/contacts/${encodeURIComponent(contactId)}/personal-information`,
    {
      method: 'PUT',
      headers: jsonHeaders(agentId),
      body: JSON.stringify(input),
    },
  )
}

export function sendMessage(
  agentId: string,
  conversationId: string,
  text: string,
): Promise<{ id: string }> {
  return request<{ id: string }>(
    `/v1/conversations/${encodeURIComponent(conversationId)}/messages`,
    {
      method: 'POST',
      headers: jsonHeaders(agentId),
      body: JSON.stringify({ text }),
    },
  )
}

export function markRead(agentId: string, conversationId: string): Promise<void> {
  return request<void>(`/v1/conversations/${encodeURIComponent(conversationId)}/messages`, {
    method: 'PATCH',
    headers: jsonHeaders(agentId),
    body: JSON.stringify({ status: 'read' }),
  })
}

export function resolveConversation(agentId: string, conversationId: string): Promise<void> {
  return request<void>(`/v1/conversations/${encodeURIComponent(conversationId)}`, {
    method: 'PATCH',
    headers: jsonHeaders(agentId),
    body: JSON.stringify({ status: 'resolved' }),
  })
}

export function simulateWebhook(payload: WebhookPayload): Promise<void> {
  return request<void>('/v1/webhooks/1channel', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  })
}
