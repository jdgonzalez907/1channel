package main

import (
	"context"
	"errors"
	"uuid"

	"github.com/jackc/pgx/v5"

	"github.com/jdgonzalez907/1channel/internal/modules/contacts"
	contactsapp "github.com/jdgonzalez907/1channel/internal/modules/contacts/app"
	contactshttp "github.com/jdgonzalez907/1channel/internal/modules/contacts/infra/http"
	contactspg "github.com/jdgonzalez907/1channel/internal/modules/contacts/infra/pg"
	convapp "github.com/jdgonzalez907/1channel/internal/modules/conversations/app"
	convhttp "github.com/jdgonzalez907/1channel/internal/modules/conversations/infra/http"
	"github.com/jdgonzalez907/1channel/internal/modules/conversations/infra/messagesender"
	convpg "github.com/jdgonzalez907/1channel/internal/modules/conversations/infra/pg"
	"github.com/jdgonzalez907/1channel/internal/modules/users"
	usersapp "github.com/jdgonzalez907/1channel/internal/modules/users/app"
	usershttp "github.com/jdgonzalez907/1channel/internal/modules/users/infra/http"
	userspg "github.com/jdgonzalez907/1channel/internal/modules/users/infra/pg"
	sharedmiddleware "github.com/jdgonzalez907/1channel/internal/shared/infra/http/middleware"
	"github.com/jdgonzalez907/1channel/internal/shared/infra/pgdb"
)

type dependencies struct {
	userWriteHandler                *usershttp.UserWriteHandler
	userReadHandler                 *usershttp.UserReadHandler
	conversationWriteHandler        *convhttp.ConversationWriteHandler
	contactWebhookHandler           *convhttp.ContactWebhookHandler
	conversationReadHandler         *convhttp.ConversationReadHandler
	contactReadHandler              *contactshttp.ContactReadHandler
	personalInformationWriteHandler *contactshttp.PersonalInformationWriteHandler
	personalInformationReadHandler  *contactshttp.PersonalInformationReadHandler
	userLookup                      sharedmiddleware.UserLookup
}

func newDependencies(db *pgdb.DB) dependencies {
	usersAPI, userWriteHandler := newUsersModule(db)
	contactsAPI, personalInformationWriteHandler, personalInformationReadHandler := newContactsModule(db)
	conversationWriteHandler, contactWebhookHandler := newConversationsModule(db, usersAPI, contactsAPI)

	return dependencies{
		userWriteHandler:                userWriteHandler,
		userReadHandler:                 usershttp.NewUserReadHandler(db.Queries),
		conversationWriteHandler:        conversationWriteHandler,
		contactWebhookHandler:           contactWebhookHandler,
		conversationReadHandler:         convhttp.NewConversationReadHandler(db.Queries),
		contactReadHandler:              contactshttp.NewContactReadHandler(db.Queries),
		personalInformationWriteHandler: personalInformationWriteHandler,
		personalInformationReadHandler:  personalInformationReadHandler,
		userLookup:                      newUserLookup(db),
	}
}

func newUserLookup(db *pgdb.DB) sharedmiddleware.UserLookup {
	return func(ctx context.Context, id uuid.UUID) error {
		if _, err := db.Queries.FindUserByID(ctx, pgdb.UUID(id)); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return users.ErrUserNotFound
			}

			return err
		}

		return nil
	}
}

func newUsersModule(db *pgdb.DB) (users.UsersAPI, *usershttp.UserWriteHandler) {
	userRepository := userspg.NewUserRepository(db)
	findUserByID := usersapp.NewFindUserByID(userRepository)
	createUser := usersapp.NewCreateUser(userRepository)

	return users.NewUsersAPI(findUserByID, createUser), usershttp.NewUserWriteHandler(createUser)
}

func newContactsModule(db *pgdb.DB) (contacts.ContactsAPI, *contactshttp.PersonalInformationWriteHandler, *contactshttp.PersonalInformationReadHandler) {
	contactRepository := contactspg.NewContactRepository(db)
	personalInformationRepository := contactspg.NewPersonalInformationRepository(db)

	contactsAPI := contacts.NewContactsAPI(
		contactsapp.NewGetOrCreateContactByExternalID(contactRepository),
		contactsapp.NewFindContactByID(contactRepository),
	)

	registerContactPersonalInformation := contactsapp.NewRegisterContactPersonalInformation(personalInformationRepository, contactRepository)

	writeHandler := contactshttp.NewPersonalInformationWriteHandler(registerContactPersonalInformation)
	readHandler := contactshttp.NewPersonalInformationReadHandler(db.Queries)

	return contactsAPI, writeHandler, readHandler
}

func newConversationsModule(db *pgdb.DB, usersAPI users.UsersAPI, contactsAPI contacts.ContactsAPI) (*convhttp.ConversationWriteHandler, *convhttp.ContactWebhookHandler) {
	conversationRepository := convpg.NewConversationRepository(db)
	sendAgentMessage := convapp.NewSendAgentMessage(conversationRepository, usersAPI, contactsAPI, messagesender.NewStubSender())
	agentReadConversation := convapp.NewAgentReadConversation(conversationRepository, usersAPI)
	resolveConversation := convapp.NewResolveConversation(conversationRepository, usersAPI)
	receiveContactMessage := convapp.NewReceiveContactMessage(conversationRepository, contactsAPI)
	receiveContactMessageEdit := convapp.NewReceiveContactMessageEdit(conversationRepository, contactsAPI)
	receiveContactMessageDelete := convapp.NewReceiveContactMessageDelete(conversationRepository, contactsAPI)
	receiveContactMessageRead := convapp.NewReceiveContactMessageRead(conversationRepository, contactsAPI)

	writeHandler := convhttp.NewConversationWriteHandler(sendAgentMessage, agentReadConversation, resolveConversation)
	webhookHandler := convhttp.NewContactWebhookHandler(
		receiveContactMessage,
		receiveContactMessageEdit,
		receiveContactMessageDelete,
		receiveContactMessageRead,
	)

	return writeHandler, webhookHandler
}
