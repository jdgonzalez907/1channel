package app

import (
	"context"
	"errors"
	"uuid"

	"github.com/jdgonzalez907/1channel/internal/modules/users/domain"
)

var ErrFindingUserByID = errors.New("finding user by id")

type FindUserByIDInput struct {
	ID uuid.UUID
}

type FindUserByID interface {
	Execute(ctx context.Context, input FindUserByIDInput) (*domain.User, error)
}

type findUserByID struct {
	userRepository domain.UserRepository
}

func NewFindUserByID(userRepository domain.UserRepository) FindUserByID {
	return &findUserByID{userRepository: userRepository}
}

func (uc *findUserByID) Execute(ctx context.Context, input FindUserByIDInput) (*domain.User, error) {
	user, err := uc.userRepository.FindByID(ctx, input.ID)
	if err != nil {
		return nil, uc.joinErr(err)
	}

	if user == nil {
		return nil, uc.joinErr(domain.ErrUserNotFound)
	}

	return user, nil
}

func (uc *findUserByID) joinErr(errs ...error) error {
	return errors.Join(ErrFindingUserByID, errors.Join(errs...))
}
