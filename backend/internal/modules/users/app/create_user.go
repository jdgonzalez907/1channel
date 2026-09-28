package app

import (
	"context"
	"errors"
	"time"
	"uuid"

	"github.com/jdgonzalez907/1channel/internal/modules/users/domain"
)

var ErrCreatingUser = errors.New("creating user failed")

type CreateUserInput struct {
	ID        uuid.UUID
	CreatedAt time.Time
}

type CreateUser interface {
	Execute(ctx context.Context, input CreateUserInput) (*domain.User, error)
}

type createUser struct {
	userRepository domain.UserRepository
}

func NewCreateUser(userRepository domain.UserRepository) CreateUser {
	return &createUser{userRepository: userRepository}
}

func (uc *createUser) Execute(ctx context.Context, input CreateUserInput) (*domain.User, error) {
	user, err := domain.NewUser(input.ID, input.CreatedAt)
	if err != nil {
		return nil, uc.joinErr(err)
	}

	if err := uc.userRepository.Save(ctx, user); err != nil {
		return nil, uc.joinErr(err)
	}

	return user, nil
}

func (uc *createUser) joinErr(errs ...error) error {
	return errors.Join(ErrCreatingUser, errors.Join(errs...))
}
