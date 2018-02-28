package database

import (
	"fmt"

	"github.com/fireeye/gocrack/server/authentication"
	"github.com/fireeye/gocrack/server/storage"

	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

func init() {
	authentication.Register("database", &DatabaseAuthPlugin{})
}

// DatabaseAuthPlugin implements Open which is used to register the GoCrack database authentication provider with the backend
type DatabaseAuthPlugin struct{}

// Open initializes the GoCrack database Authentication Provider
func (s *DatabaseAuthPlugin) Open(db authentication.AuthStorageBackend, cfg authentication.PluginSettings) (authentication.AuthAPI, error) {
	return Init(db, cfg)
}

const (
	// DefaultUsername is the first user created in the system if no users exist
	DefaultUsername = "admin"
	// DefaultPassword is the default password for the DefaultUsername
	DefaultPassword = "ch@ng3me!"
	// unexported fields
	errWrongType = "authentication.backend_settings.%s is of the wrong type. Expected %s, got %T"
)

func convertRawConfigToConfig(in map[string]interface{}) (out *Config, err error) {
	out = &Config{}

	if val, ok := in["bcrypt_cost"]; ok {
		i, ok := val.(int)
		if !ok {
			return nil, fmt.Errorf(errWrongType, "bcrypt_cost", "int", val)
		}
		out.BCryptCost = i
	}

	if val, ok := in["allow_registration"]; ok {
		i, ok := val.(bool)
		if !ok {
			return nil, fmt.Errorf(errWrongType, "allow_registration", "bool", val)
		}
		out.AllowRegistration = i
	}

	if out.BCryptCost < bcrypt.MinCost || out.BCryptCost > bcrypt.MaxCost {
		out.BCryptCost = bcrypt.DefaultCost
	}
	return out, nil
}

type (
	// Config describes all the configuration options for a database backed auth
	Config struct {
		BCryptCost        int
		AllowRegistration bool
	}

	// DatabaseAuth is an authentication backend powered by one of GoCrack's storage implementations
	DatabaseAuth struct {
		db  authentication.AuthStorageBackend
		cfg Config
	}
)

// Init creates the authentication backend
func Init(db authentication.AuthStorageBackend, cfg authentication.PluginSettings) (*DatabaseAuth, error) {
	dcfg, err := convertRawConfigToConfig(cfg)
	if err != nil {
		return nil, err
	}

	authBackend := &DatabaseAuth{
		db:  db,
		cfg: *dcfg,
	}

	// Check and see if there are any users first
	if hasUsers, err := checkForUsers(db); err != nil {
		return nil, err
	} else if !hasUsers {
		log.Warn().Str("username", DefaultUsername).Str("password", DefaultPassword).Msg("Created default admin account")
		if err := authBackend.CreateUser(storage.User{
			Username:    DefaultUsername,
			Password:    DefaultPassword,
			IsSuperUser: true,
		}); err != nil {
			return nil, err
		}
	}

	return authBackend, nil
}

// checkForUsers returns true if a user exists in the system.
func checkForUsers(db authentication.AuthStorageBackend) (bool, error) {
	users, err := db.GetUsers()
	if err != nil {
		return false, err
	}
	return len(users) > 0, nil
}

// Login searches the database backend for a matching user record
func (s *DatabaseAuth) Login(username, password string) (*storage.User, error) {
	return s.db.SearchForUserByPassword(username, func(passwordFromDb string) bool {
		if err := bcrypt.CompareHashAndPassword([]byte(passwordFromDb), []byte(password)); err != nil {
			return false
		}
		return true
	})
}

// CreateUser creates a new user in the database
func (s *DatabaseAuth) CreateUser(user storage.User) error {
	securePassword, err := s.GenerateSecurePassword(user.Password)
	if err != nil {
		return err
	}

	user.Password = securePassword
	return s.db.CreateUser(&user)
}

// UserCanChangePassword indicates that the user can change their password with this backend
func (s *DatabaseAuth) UserCanChangePassword() bool {
	return true
}

// CanUsersRegister indicates if the GoCrack administrator has allowed new user registration
func (s *DatabaseAuth) CanUsersRegister() bool {
	return s.cfg.AllowRegistration
}

// GenerateSecurePassword generates a cryptographically secure password string from a plaintext password
func (s *DatabaseAuth) GenerateSecurePassword(password string) (string, error) {
	if password == "" {
		return "", authentication.ErrPasswordEmpty
	}

	if !authentication.CheckPasswordRequirement(password) {
		return "", authentication.ErrFailsRequirements
	}

	b, err := bcrypt.GenerateFromPassword([]byte(password), s.cfg.BCryptCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
