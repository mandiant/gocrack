package ldap

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/fireeye/gocrack/server/authentication"
	"github.com/fireeye/gocrack/server/storage"

	uuid "github.com/satori/go.uuid"
	ldap "gopkg.in/ldap.v2"
)

func init() {
	authentication.Register("ldap", &LDAPAuthPlugin{})
}

type LDAPAuthPlugin struct{}

func (s *LDAPAuthPlugin) Open(db authentication.AuthStorageBackend, cfg authentication.PluginSettings) (authentication.AuthAPI, error) {
	return Init(db, cfg)
}

var (
	ErrUserNotFound = errors.New("user not found")
	ErrMoreThanOne  = errors.New("more than one record has been found")
	ErrInvalidCert  = errors.New("invalid root ca cert")
	ErrDisabled     = errors.New("disabled in ldap authentication")
)

func checkForString(in map[string]interface{}, expectedToHave string) (string, error) {
	var out string

	val, ok := in[expectedToHave]
	if ok {
		if out, ok = in[expectedToHave].(string); !ok {
			return "", fmt.Errorf("authentication.backend_settings.%s must be a string", expectedToHave)
		}
	}

	if val == "" {
		return "", fmt.Errorf("authentication.backend_settings.%s must not be empty", expectedToHave)
	}
	return out, nil
}

func convertRawConfigToConfig(in map[string]interface{}) (*Options, error) {
	var cfg Options
	var err error

	if cfg.Address, err = checkForString(in, "address"); err != nil {
		return nil, err
	}

	if cfg.Base, err = checkForString(in, "base_dn"); err != nil {
		return nil, err
	}

	if cfg.BindDN, err = checkForString(in, "bind_dn"); err != nil {
		return nil, err
	}

	if cfg.BindPassword, err = checkForString(in, "bind_password"); err != nil {
		return nil, err
	}

	if cfg.RootCACert, err = checkForString(in, "root_ca"); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Options contains all the LDAP configuration options
type Options struct {
	Address      string `yaml:"address"`
	Base         string `yaml:"base_dn"`
	BindDN       string `yaml:"bind_dn"`
	BindPassword string `yaml:"bind_password"`
	RootCACert   string `yaml:"root_ca"`
}

// Backend is an authentication backend that queries an LDAP/Active Directory server for authentication
type Backend struct {
	// unexported fields below
	c     *ldap.Conn
	mu    *sync.Mutex
	certp *x509.CertPool
	db    authentication.AuthStorageBackend
	*Options
}

// Init creates a new LDAP authentication backend
func Init(db authentication.AuthStorageBackend, cfg authentication.PluginSettings) (*Backend, error) {
	var pool *x509.CertPool

	rcfg, err := convertRawConfigToConfig(cfg)
	if err != nil {
		return nil, err
	}

	if rcfg.RootCACert != "" {
		block, _ := pem.Decode([]byte(rcfg.RootCACert))
		if block == nil {
			return nil, ErrInvalidCert
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, ErrInvalidCert
		}

		pool = x509.NewCertPool()
		pool.AddCert(cert)
	}

	return &Backend{
		mu:      &sync.Mutex{},
		Options: rcfg,
		certp:   pool,
		db:      db,
	}, nil
}

// Close the LDAP client if it's open
func (s *Backend) Close() {
	if s.c != nil {
		s.c.Close()
		s.c = nil
	}
	return
}

func (s *Backend) connect() error {
	var hostname = s.Address
	// connection has already been established
	if s.c != nil {
		return nil
	}

	if strings.Contains(hostname, ":") {
		hostname = s.Address[:strings.Index(hostname, ":")]
	}

	l, err := ldap.DialTLS("tcp", s.Address, &tls.Config{
		ServerName: hostname,
		RootCAs:    s.certp,
	})

	if err != nil {
		return err
	}

	s.c = l
	return nil
}

func (s *Backend) getUser(username string) (props map[string]string, err error) {
	if err = s.connect(); err != nil {
		return
	}

	if err = s.c.Bind(s.BindDN, s.BindPassword); err != nil {
		return
	}

	attribs := []string{"givenName", "sn", "mail", "uid"}
	searchRequest := ldap.NewSearchRequest(
		s.Base,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(uid=%s)", username),
		attribs,
		nil,
	)

	sr, err := s.c.Search(searchRequest)
	if err != nil {
		return
	}

	if len(sr.Entries) == 0 {
		err = ErrUserNotFound
		return
	}

	if len(sr.Entries) > 1 {
		err = ErrMoreThanOne
		return
	}

	userDN := sr.Entries[0].DN
	props = make(map[string]string)
	for _, attr := range attribs {
		props[attr] = sr.Entries[0].GetAttributeValue(attr)
	}
	props["dn"] = userDN

	return
}

// Login searches the database backend for a matching user record
func (s *Backend) Login(username, password string) (*storage.User, error) {
	if password == "" {
		return nil, authentication.ErrPasswordEmpty
	}

	props, err := s.getUser(username)
	if err != nil {
		return nil, err
	}

	// Bind as the user to verify their password
	if err = s.c.Bind(props["dn"], password); err != nil {
		return nil, err
	}

	// User is valid (according to ldap)!
	// check and see if we have a record in the storage backend for them
	rec, err := s.db.SearchForUserByPassword(username, func(passwordFromDb string) bool {
		// this should always be user_is_ldap for entries created by the LDAP provider
		if passwordFromDb == "user_is_ldap" {
			return true
		}
		return false
	})

	// Create the user if the record is not found
	if err == storage.ErrNotFound {
		// XXX(cschmitt): this isnt ideal, we should automatically make them super user if they're in a specific OU
		isUserFirstAndShouldBeAdmin, err := s.shouldUserBeAdmin()
		if err != nil {
			return nil, err
		}

		rec = &storage.User{
			Username:     username,
			IsSuperUser:  isUserFirstAndShouldBeAdmin,
			EmailAddress: props["mail"],
			UserUUID:     uuid.NewV4().String(),
			Password:     "user_is_ldap",
		}

		if err = s.db.CreateUser(rec); err != nil {
			return nil, err
		}
	}
	return rec, nil
}

// shouldUserBeAdmin will return true if no users exist in the system as an administrator.. there should always be one!
func (s *Backend) shouldUserBeAdmin() (bool, error) {
	users, err := s.db.GetUsers()
	if err != nil {
		return false, err
	}

	for _, user := range users {
		if user.IsSuperUser {
			return false, nil
		}
	}
	return true, nil
}

// CreateUser is disabled in the LDAP authentication backend
func (s *Backend) CreateUser(user storage.User) error {
	return ErrDisabled
}

// UserCanChangePassword is disabled in the LDAP authentication backend
func (s *Backend) UserCanChangePassword() bool {
	return false
}

// CanUsersRegister is disabled in the LDAP authentication backend
func (s *Backend) CanUsersRegister() bool {
	return false
}

// GenerateSecurePassword is disabled in the LDAP authentication backend
func (s *Backend) GenerateSecurePassword(password string) (string, error) {
	return "", ErrDisabled
}
