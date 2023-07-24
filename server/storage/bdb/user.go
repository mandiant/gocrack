package bdb

import (
	"time"

	"github.com/mandiant/gocrack/server/storage"

	"github.com/asdine/storm"
	"github.com/google/uuid"
)

// SearchForUserByPassword locates the user record by username. If a record is found, the checker function will be called to validate the password.
func (s *BoltBackend) SearchForUserByPassword(username string, checker storage.PasswordCheckFunc) (*storage.User, error) {
	var bu boltUser
	if err := s.db.From("users").One("Username", username, &bu); err != nil {
		return nil, convertErr(err)
	}

	if !checker(bu.Password) {
		return nil, convertErr(storm.ErrNotFound)
	}

	return &bu.User, nil
}

// CreateUser saves the record into the database
func (s *BoltBackend) CreateUser(user *storage.User) (err error) {
	user.CreatedAt = time.Now().UTC()
	if user.UserUUID == "" {
		user.UserUUID = uuid.NewString()
	}

	if err = s.db.From("users").Save(&boltUser{User: *user, DocVersion: curUserVer}); err != nil {
		err = convertErr(err)
	}
	return err
}

func getUserFromNode(n storm.Node, userUUID string) (*storage.User, error) {
	var bu boltUser
	if err := n.From("users").One("UserUUID", userUUID, &bu); err != nil {
		return nil, convertErr(err)
	}
	return &bu.User, nil
}

// GetUserByID returns a user record given the users unique uuid.
func (s *BoltBackend) GetUserByID(userUUID string) (*storage.User, error) {
	return getUserFromNode(s.db, userUUID)
}

// GetUsers returns a list of all users within the GoCrack system
func (s *BoltBackend) GetUsers() ([]storage.User, error) {
	var users []boltUser
	if err := s.db.From("users").All(&users); err != nil {
		return nil, convertErr(err)
	}

	susers := make([]storage.User, len(users))
	for i, user := range users {
		susers[i] = user.User
	}
	return susers, nil
}

func (s *BoltBackend) EditUser(userUUID string, req storage.UserModifyRequest) error {
	txn, err := s.db.From("users").Begin(true)
	if err != nil {
		return err
	}
	defer txn.Rollback()

	var tmp boltUser
	var updated bool

	if err = txn.One("UserUUID", userUUID, &tmp); err != nil {
		return convertErr(err)
	}

	if req.UserIsAdmin != nil && *req.UserIsAdmin != tmp.IsSuperUser {
		tmp.IsSuperUser = *req.UserIsAdmin
		updated = true
	}

	if req.Email != nil && *req.Email != tmp.EmailAddress {
		tmp.EmailAddress = *req.Email
		updated = true
	}

	if req.Password != nil && *req.Password != tmp.Password {
		tmp.Password = *req.Password
		updated = true
	}

	if updated {
		if err = txn.Update(&tmp); err != nil {
			return err
		}
		txn.Commit()
	}

	return nil
}
