export type ConversationStatus = 'pending' | 'assigned' | 'expired' | 'resolved'

export type ConversationTab = 'open' | 'finished'

export type MessageStatus = 'sent' | 'read' | 'deleted' | 'failed'

export type MessageType = 'text'

export type MessageOwner = 'agent' | 'contact'

export type WebhookEvent =
  | 'message.received'
  | 'message.edited'
  | 'message.deleted'
  | 'message.read'

export interface ContactRef {
  id: string
  external_id: string
  label: string
}

export interface PersonalInformation {
  id: string
  identification_number: string
  first_name: string | null
  last_name: string | null
  phone_number: string | null
  email: string | null
  address: string | null
  created_at: string
  updated_at: string
}

export interface SavePersonalInformationInput {
  identification_number: string
  first_name: string
  last_name: string
  phone_number: string
  email: string
  address: string
}

export interface LastMessage {
  text: string | null
  sent_at: string
  owner: MessageOwner
}

export interface ConversationListItem {
  id: string
  status: ConversationStatus
  contact: ContactRef
  last_message: LastMessage
  unread_count: number
}

export interface ConversationListResponse {
  items: ConversationListItem[]
  next_before_sent_at: string | null
  next_before_id: string | null
}

export interface Message {
  id: string
  status: MessageStatus
  type: MessageType
  text: string | null
  owner: MessageOwner
  sent_at: string
  read_at: string | null
  edited_at: string | null
  deleted_at: string | null
}

export interface ConversationDetail {
  id: string
  status: ConversationStatus
  contact: ContactRef
  personal_information: PersonalInformation | null
  agent_id: string | null
  unread_count: number
  messages: Message[]
  next_before_sent_at: string | null
  next_before_id: string | null
}

export interface Contact {
  id: string
  external_id: string
  label: string
  display_name: string | null
  personal_information: PersonalInformation | null
  created_at: string
}

export interface User {
  id: string
  created_at: string
}

export interface ProblemDetails {
  title: string
  status: number
  detail: string
  instance?: string
}

export interface PaginationPosition {
  before_sent_at?: string
  before_id?: string
}

export interface WebhookPayload {
  event: WebhookEvent
  external_contact_id: string
  external_message_id: string
  display_name?: string
  text?: string
}
