package main

import (
	"github.com/jdgonzalez907/1channel/internal/modules/contacts"
	contactsapp "github.com/jdgonzalez907/1channel/internal/modules/contacts/app"
	contactspg "github.com/jdgonzalez907/1channel/internal/modules/contacts/infra/pg"
	convapp "github.com/jdgonzalez907/1channel/internal/modules/conversations/app"
	convhttp "github.com/jdgonzalez907/1channel/internal/modules/conversations/infra/http"
	"github.com/jdgonzalez907/1channel/internal/modules/conversations/infra/messagesender"
	convpg "github.com/jdgonzalez907/1channel/internal/modules/conversations/infra/pg"
	"github.com/jdgonzalez907/1channel/internal/modules/users"
	usersapp "github.com/jdgonzalez907/1channel/internal/modules/users/app"
	usershttp "github.com/jdgonzalez907/1channel/internal/modules/users/infra/http"
	userspg "github.com/jdgonzalez907/1channel/internal/modules/users/infra/pg"
	"github.com/jdgonzalez907/1channel/internal/shared/infra/pgdb"
)

type dependencies struct {
	usersHandler         *usershttp.UserHandler
	conversationsHandler *convhttp.ConversationHandler
}

func newDependencies(db *pgdb.DB) dependencies {
	usersAPI, usersHandler := newUsersModule(db)
	contactsAPI := newContactsModule(db)
	conversationsHandler := newConversationsModule(db, usersAPI, contactsAPI)

	return dependencies{
		usersHandler:         usersHandler,
		conversationsHandler: conversationsHandler,
	}
}

func newUsersModule(db *pgdb.DB) (users.UsersAPI, *usershttp.UserHandler) {
	userRepository := userspg.NewUserRepository(db)
	findUserByID := usersapp.NewFindUserByID(userRepository)
	createUser := usersapp.NewCreateUser(userRepository)

	return users.NewUsersAPI(findUserByID, createUser), usershttp.NewUserHandler(createUser)
}

func newContactsModule(db *pgdb.DB) contacts.ContactsAPI {
	contactRepository := contactspg.NewContactRepository(db)

	return contacts.NewContactsAPI(
		contactsapp.NewGetOrCreateContactByExternalID(contactRepository),
		contactsapp.NewFindContactByID(contactRepository),
	)
}

func newConversationsModule(db *pgdb.DB, usersAPI users.UsersAPI, contactsAPI contacts.ContactsAPI) *convhttp.ConversationHandler {
	conversationRepository := convpg.NewConversationRepository(db)
	sendAgentMessage := convapp.NewSendAgentMessage(conversationRepository, usersAPI, contactsAPI, messagesender.NewStubSender())
	agentReadConversation := convapp.NewAgentReadConversation(conversationRepository, usersAPI)
	resolveConversation := convapp.NewResolveConversation(conversationRepository, usersAPI)

	return convhttp.NewConversationHandler(sendAgentMessage, agentReadConversation, resolveConversation)
}
