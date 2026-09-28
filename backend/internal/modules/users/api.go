package users

import (
	"context"
	"time"
	"uuid"

	"github.com/jdgonzalez907/1channel/internal/modules/users/app"
	"github.com/jdgonzalez907/1channel/internal/modules/users/domain"
)

var ErrUserNotFound = domain.ErrUserNotFound

type UsersAPI interface {
	FindUserByID(ctx context.Context, id uuid.UUID) (uuid.UUID, error)
	CreateUser(ctx context.Context, input CreateUserInput) (uuid.UUID, error)
}

type CreateUserInput struct {
	ID        uuid.UUID
	CreatedAt time.Time
}

type usersAPI struct {
	findUserByID app.FindUserByID
	createUser   app.CreateUser
}

func NewUsersAPI(findUserByID app.FindUserByID, createUser app.CreateUser) UsersAPI {
	return &usersAPI{findUserByID: findUserByID, createUser: createUser}
}

func (a *usersAPI) FindUserByID(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	user, err := a.findUserByID.Execute(ctx, app.FindUserByIDInput{ID: id})
	if err != nil {
		return uuid.Nil(), err
	}

	return user.ID(), nil
}

func (a *usersAPI) CreateUser(ctx context.Context, input CreateUserInput) (uuid.UUID, error) {
	user, err := a.createUser.Execute(ctx, app.CreateUserInput{ID: input.ID, CreatedAt: input.CreatedAt})
	if err != nil {
		return uuid.Nil(), err
	}

	return user.ID(), nil
}
