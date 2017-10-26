package test

import "github.com/fireeye/gocrack/server/storage"

type FakeAuthPlugin struct {
	db *FakeDatabase
}

func NewFakeAuthProv(db *FakeDatabase) *FakeAuthPlugin {
	return &FakeAuthPlugin{
		db: db,
	}
}

func (s *FakeAuthPlugin) Login(username, password string) (user *storage.User, err error) {
	return s.db.SearchForUserByPassword(username, func(passwordFromDb string) bool {
		return password == passwordFromDb
	})
}

func (s *FakeAuthPlugin) CreateUser(user storage.User) error {
	return s.db.CreateUser(&user)
}

func (s *FakeAuthPlugin) UserCanChangePassword() bool {
	return false
}

func (s *FakeAuthPlugin) CanUsersRegister() bool {
	return false
}

func (s *FakeAuthPlugin) GenerateSecurePassword(password string) (string, error) {
	return password, nil
}
