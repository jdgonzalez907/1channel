package http

import (
	"net/http"
	"time"
	"uuid"

	"github.com/go-chi/chi/v5"

	"github.com/jdgonzalez907/1channel/internal/modules/conversations/domain"
	"github.com/jdgonzalez907/1channel/internal/shared/infra/http/httperror"
	"github.com/jdgonzalez907/1channel/internal/shared/infra/http/httputil"
	"github.com/jdgonzalez907/1channel/internal/shared/infra/http/middleware"
	"github.com/jdgonzalez907/1channel/internal/shared/infra/pgdb"
	"github.com/jdgonzalez907/1channel/internal/shared/infra/pgdb/sqlc"
)

const readPageSize = 20

const (
	queryParamStatus            = "status"
	queryParamExternalContactID = "external_contact_id"
	queryParamBeforeSentAt      = "before_sent_at"
	queryParamBeforeID          = "before_id"

	tabOpen     = "open"
	tabFinished = "finished"

	messageOwnerAgent   = "agent"
	messageOwnerContact = "contact"
)

type ConversationReadHandler struct {
	queries *sqlc.Queries
}

func NewConversationReadHandler(queries *sqlc.Queries) *ConversationReadHandler {
	return &ConversationReadHandler{queries: queries}
}

func (h *ConversationReadHandler) Register(r chi.Router) {
	r.Get("/conversations", h.handleListConversations)
	r.Get("/conversations/{id}", h.handleGetConversation)
}

func (h *ConversationReadHandler) handleListConversations(w http.ResponseWriter, r *http.Request) {
	agentID, ok := middleware.RequireUserID(w, r)
	if !ok {
		return
	}

	statuses, ok := statusesFromTab(r.URL.Query().Get(queryParamStatus))
	if !ok {
		httperror.Unprocessable(w, r, "status must be open or finished")
		return
	}

	beforeSentAt, beforeID, ok := beforePositionFromQuery(w, r)
	if !ok {
		return
	}

	contactID, ok := h.resolveContactFilter(w, r)
	if !ok {
		return
	}

	rows, err := h.queries.ListConversationsForAgent(r.Context(), sqlc.ListConversationsForAgentParams{
		AgentID:      pgdb.UUID(agentID),
		Statuses:     statuses,
		ContactID:    pgdb.UUIDPtr(contactID),
		BeforeSentAt: pgdb.TimestampPtr(beforeSentAt),
		BeforeID:     pgdb.UUIDPtr(beforeID),
		PageSize:     int32(readPageSize),
	})
	if err != nil {
		httperror.Generic(w, r, err)
		return
	}

	rows, hasMore := page(rows)

	var nextBeforeSentAt *string
	var nextBeforeID *string
	if hasMore {
		last := rows[len(rows)-1]
		nextBeforeSentAt, nextBeforeID = nextBeforePosition(pgdb.FromTimestamp(last.LastMessageAt), pgdb.FromUUID(last.ID))
	}

	items := make([]ConversationListItemResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, conversationListItemFromRow(row))
	}

	httputil.JSON(w, http.StatusOK, ConversationListResponse{
		Items:            items,
		NextBeforeSentAt: nextBeforeSentAt,
		NextBeforeID:     nextBeforeID,
	})
}

func (h *ConversationReadHandler) handleGetConversation(w http.ResponseWriter, r *http.Request) {
	conversationID, ok := httputil.RequirePathUUID(w, r, "id", "invalid conversation id")
	if !ok {
		return
	}

	agentID, ok := middleware.RequireUserID(w, r)
	if !ok {
		return
	}

	beforeSentAt, beforeID, ok := beforePositionFromQuery(w, r)
	if !ok {
		return
	}

	conversation, err := h.queries.FindConversationWithContactByID(r.Context(), pgdb.UUID(conversationID))
	if err != nil {
		httputil.LookupError(w, r, err, "conversation not found")
		return
	}

	if !isConversationVisible(conversation, agentID) {
		httperror.Forbidden(w, r, "conversation is not accessible")
		return
	}

	rows, err := h.queries.ListConversationMessagesPage(r.Context(), sqlc.ListConversationMessagesPageParams{
		ConversationID: conversation.ID,
		BeforeSentAt:   pgdb.TimestampPtr(beforeSentAt),
		BeforeID:       pgdb.UUIDPtr(beforeID),
		PageSize:       int32(readPageSize),
	})
	if err != nil {
		httperror.Generic(w, r, err)
		return
	}

	rows, hasMore := page(rows)

	var nextBeforeSentAt *string
	var nextBeforeID *string
	if hasMore {
		oldest := rows[len(rows)-1]
		nextBeforeSentAt, nextBeforeID = nextBeforePosition(pgdb.FromTimestamp(oldest.SentAt), pgdb.FromUUID(oldest.ID))
	}

	messages := make([]ConversationMessageResponse, 0, len(rows))
	for i := len(rows) - 1; i >= 0; i-- {
		messages = append(messages, conversationMessageFromRow(rows[i]))
	}

	httputil.JSON(w, http.StatusOK, ConversationDetailResponse{
		ID:     pgdb.UUIDString(conversation.ID),
		Status: conversation.Status,
		Contact: ContactResponse{
			ID:         pgdb.UUIDString(conversation.ContactID),
			ExternalID: conversation.ExternalContactID,
			Label: httputil.ContactLabel(
				conversation.FirstName,
				conversation.LastName,
				conversation.DisplayName,
				conversation.ExternalContactID,
			),
		},
		PersonalInformation: personalInformationFromConversationRow(conversation),
		AgentID:             pgdb.UUIDStringPtr(conversation.UserID),
		UnreadCount:         conversation.UnreadCount,
		Messages:            messages,
		NextBeforeSentAt:    nextBeforeSentAt,
		NextBeforeID:        nextBeforeID,
	})
}

func (h *ConversationReadHandler) resolveContactFilter(w http.ResponseWriter, r *http.Request) (*uuid.UUID, bool) {
	externalContactID := r.URL.Query().Get(queryParamExternalContactID)
	if externalContactID == "" {
		return nil, true
	}

	contact, err := h.queries.FindContactByExternalContactID(r.Context(), externalContactID)
	if err != nil {
		httputil.LookupError(w, r, err, "contact not found")
		return nil, false
	}

	id := pgdb.FromUUID(contact.ID)

	return &id, true
}

func statusesFromTab(status string) ([]string, bool) {
	switch status {
	case tabOpen:
		return []string{
			domain.ConversationStatusPending.String(),
			domain.ConversationStatusAssigned.String(),
		}, true
	case tabFinished:
		return []string{
			domain.ConversationStatusResolved.String(),
			domain.ConversationStatusExpired.String(),
		}, true
	default:
		return nil, false
	}
}

func beforePositionFromQuery(w http.ResponseWriter, r *http.Request) (*time.Time, *uuid.UUID, bool) {
	query := r.URL.Query()
	rawSentAt := query.Get(queryParamBeforeSentAt)
	rawID := query.Get(queryParamBeforeID)

	if rawSentAt == "" && rawID == "" {
		return nil, nil, true
	}

	if rawSentAt == "" || rawID == "" {
		httperror.BadRequest(w, r, "before_sent_at and before_id must be provided together")
		return nil, nil, false
	}

	at, err := time.Parse(time.RFC3339Nano, rawSentAt)
	if err != nil {
		httperror.BadRequest(w, r, "invalid before_sent_at")
		return nil, nil, false
	}

	id, err := uuid.Parse(rawID)
	if err != nil {
		httperror.BadRequest(w, r, "invalid before_id")
		return nil, nil, false
	}

	return &at, &id, true
}

func isConversationVisible(conversation sqlc.FindConversationWithContactByIDRow, agentID uuid.UUID) bool {
	if !conversation.UserID.Valid {
		return true
	}

	return pgdb.FromUUID(conversation.UserID) == agentID
}

func conversationListItemFromRow(row sqlc.ListConversationsForAgentRow) ConversationListItemResponse {
	var lastMessageText *string
	if row.LastMessageStatus != domain.MessageStatusDeleted.String() {
		lastMessageText = row.LastMessageText
	}

	return ConversationListItemResponse{
		ID:     pgdb.UUIDString(row.ID),
		Status: row.Status,
		Contact: ContactResponse{
			ID:         pgdb.UUIDString(row.ContactID),
			ExternalID: row.ExternalContactID,
			Label: httputil.ContactLabel(
				row.FirstName,
				row.LastName,
				row.DisplayName,
				row.ExternalContactID,
			),
		},
		LastMessage: LastMessageResponse{
			Text:   lastMessageText,
			SentAt: pgdb.FromTimestamp(row.LastMessageAt),
			Owner:  row.LastMessageOwner,
		},
		UnreadCount: row.UnreadCount,
	}
}

func personalInformationFromConversationRow(row sqlc.FindConversationWithContactByIDRow) *PersonalInformationResponse {
	if !row.PiID.Valid {
		return nil
	}

	return &PersonalInformationResponse{
		ID:                   pgdb.UUIDString(row.PiID),
		IdentificationNumber: row.IdentificationNumber,
		FirstName:            row.FirstName,
		LastName:             row.LastName,
		PhoneNumber:          row.PhoneNumber,
		Email:                row.Email,
		Address:              row.Address,
		CreatedAt:            pgdb.FromTimestamp(row.PiCreatedAt),
		UpdatedAt:            pgdb.FromTimestamp(row.PiUpdatedAt),
	}
}

func conversationMessageFromRow(row sqlc.ListConversationMessagesPageRow) ConversationMessageResponse {
	var text *string
	if row.Status != domain.MessageStatusDeleted.String() {
		text = row.Text
	}

	owner := messageOwnerContact
	if row.UserID.Valid {
		owner = messageOwnerAgent
	}

	return ConversationMessageResponse{
		ID:        pgdb.UUIDString(row.ID),
		Status:    row.Status,
		Type:      row.Type,
		Text:      text,
		Owner:     owner,
		SentAt:    pgdb.FromTimestamp(row.SentAt),
		ReadAt:    pgdb.FromTimestampPtr(row.ReadAt),
		EditedAt:  pgdb.FromTimestampPtr(row.EditedAt),
		DeletedAt: pgdb.FromTimestampPtr(row.DeletedAt),
	}
}

func page[T any](rows []T) ([]T, bool) {
	if len(rows) > readPageSize {
		return rows[:readPageSize], true
	}

	return rows, false
}

func nextBeforePosition(at time.Time, id uuid.UUID) (*string, *string) {
	sentAt := at.UTC().Format(time.RFC3339Nano)
	value := id.String()

	return &sentAt, &value
}
