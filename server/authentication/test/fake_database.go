package test

import (
	"errors"

	"github.com/fireeye/gocrack/server/storage"
)

type FakeDatabase struct {
	users []*storage.User
}

func NewFakeDatabase() *FakeDatabase {
	return &FakeDatabase{
		users: make([]*storage.User, 0),
	}
}

func (s *FakeDatabase) CreateUser(user *storage.User) error {
	s.users = append(s.users, user)
	return nil
}

func (s *FakeDatabase) SearchForUserByPassword(username string, f storage.PasswordCheckFunc) (*storage.User, error) {
	for _, user := range s.users {
		if user.Username == username {
			if valid := f(user.Password); valid {
				return user, nil
			}
		}
	}
	return nil, errors.New("not found")
}

func (s *FakeDatabase) GetUsers() ([]storage.User, error) {
	out := make([]storage.User, len(s.users))
	for i, usr := range s.users {
		out[i] = *usr
	}
	return out, nil
}
