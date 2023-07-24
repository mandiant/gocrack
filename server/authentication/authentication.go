package authentication

import (
	"errors"
	"fmt"
	"time"

	"github.com/mandiant/gocrack/server/storage"
	"github.com/mandiant/gocrack/shared"

	jose "gopkg.in/square/go-jose.v2"
	jwt "gopkg.in/square/go-jose.v2/jwt"
)

var (
	// ErrExpired indicates that token is used after expiry time indicated in exp claim.
	ErrExpired = errors.New("authentication has expired")
	// ErrFailsRequirements indicates the password fails to meet the system requirements
	ErrFailsRequirements = errors.New("password fails requirements")
	// ErrPasswordEmpty indicates the password is empty
	ErrPasswordEmpty = errors.New("password is empty")

	// DefaultTokenExpiry indicates the default duration of a JWT if the TokenExpiry in AuthSettings is nil
	DefaultTokenExpiry = shared.HumanDuration{Duration: 24 * time.Hour}

	errNoBackend           = errors.New("authentication.backend must not be empty")
	errTokenExpiryNegative = errors.New("authentication.token_expiry must be a positive duration")
	errSecretKeyBad        = errors.New("authentication.secret_key must be set to a secure string")
)

type (
	// AuthSettings describes the basic configuration options for the modula authentication backend
	AuthSettings struct {
		Backend     string                 `yaml:"backend"`
		Settings    map[string]interface{} `yaml:"backend_settings,omitempty"`
		TokenExpiry *shared.HumanDuration  `yaml:"token_expiry,omitempty"`
		SecretKey   *string                `yaml:"secret_key,omitempty"`
	}

	// AuthAPI describes the APIs that authentication backends must implement
	AuthAPI interface {
		// Login to the authentication backend with the given username and password
		Login(username, password string) (user *storage.User, err error)
		// CreateUser stores the user into the storage backend and takes any necessary actions to create the user that will function in this backend
		CreateUser(user storage.User) (err error)
		// UserCanChangePassword indicates if a user is able to change his or her's password in this backend
		UserCanChangePassword() bool
		// GenerateSecurePassword returns a secure version that is safe for long-term storage of the password passed into the function
		GenerateSecurePassword(password string) (string, error)
		// CanUsersRegister indicates if users are able to register on the system
		CanUsersRegister() bool
	}

	// AuthStorageBackend defines the APIs we need from the storage driver to implement a authentication driver
	AuthStorageBackend interface {
		CreateUser(*storage.User) error
		SearchForUserByPassword(string, storage.PasswordCheckFunc) (*storage.User, error)
		GetUsers() ([]storage.User, error)
	}

	// ProviderAPI describes the APIs available to the a service that requires authentication
	ProviderAPI interface {
		Login(username, password string, APIOnly bool) (claimStr string, err error)
		VerifyClaim(rawclaim, expectedSubject string, auds ...string) (claim *AuthClaim, err error)
	}

	// AuthClaim is a JWT claim describing metadata about an authenticated user
	AuthClaim struct {
		Username string `json:"username"`
		UserUUID string `json:"user_uuid"`
		Email    string `json:"email"`
		IsAdmin  bool   `json:"is_admin"`
		APIOnly  bool   `json:"api_only"`
		jwt.Claims
	}

	// AuthWrapper wraps the requested provider
	AuthWrapper struct {
		as  AuthSettings
		sig jose.Signer
		key []byte
		AuthAPI
	}
)

// convertError converts the jose specific errors and returns an error we control
func convertError(err error) error {
	switch err {
	case jwt.ErrExpired:
		return ErrExpired
	default:
		return err
	}
}

// WrapProvider returns an auth wrapper that is used by services like the API to perform authentication
func WrapProvider(prov AuthAPI, as AuthSettings) *AuthWrapper {
	k := []byte(*as.SecretKey)
	sig, err := jose.NewSigner(
		jose.SigningKey{
			Algorithm: jose.HS256,
			Key:       k,
		},
		(&jose.SignerOptions{}).WithType("JWT"))
	if err != nil {
		panic(err)
	}

	return &AuthWrapper{
		AuthAPI: prov,
		as:      as,
		sig:     sig,
		key:     k,
	}
}

// Validate the configuration; setting default values and returning any errors
func (s *AuthSettings) Validate() error {
	if s.Backend == "" {
		return errNoBackend
	}

	if s.TokenExpiry == nil {
		s.TokenExpiry = &DefaultTokenExpiry
	} else if s.TokenExpiry.Duration.Nanoseconds() < 0 {
		return errTokenExpiryNegative
	}

	if s.SecretKey == nil || *s.SecretKey == "" {
		return errSecretKeyBad
	}

	return nil
}

// Login the user and set the JWT token to the header
func (s *AuthWrapper) Login(username, password string, APIOnly bool) (string, error) {
	found, err := s.AuthAPI.Login(username, password)
	if found == nil || err != nil {
		return "", convertError(err)
	}

	now := time.Now().UTC()
	claim := AuthClaim{
		Username: found.Username,
		Email:    found.EmailAddress,
		IsAdmin:  found.IsSuperUser,
		UserUUID: found.UserUUID,
		APIOnly:  APIOnly,
		Claims: jwt.Claims{
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			Expiry:    jwt.NewNumericDate(now.Add(s.as.TokenExpiry.Duration)),
			Subject:   "gocrack",
			Audience:  jwt.Audience{"api", "realtime"},
		},
	}

	return jwt.Signed(s.sig).Claims(claim).CompactSerialize()
}

// VerifyClaim parses a raw JWT claim and validates it
func (s *AuthWrapper) VerifyClaim(rawclaim, expSubj string, expAuds ...string) (*AuthClaim, error) {
	tok, err := jwt.ParseSigned(rawclaim)
	if err != nil {
		return nil, convertError(err)
	}

	claim := &AuthClaim{}
	if err = tok.Claims(s.key, claim); err != nil {
		return nil, convertError(err)
	}

	// ensure the claim has at least the proper audiences needed.
	// we dont use the library for this check because it wants *all* the audiences to exist
	for _, aud := range expAuds {
		if !claim.Audience.Contains(aud) {
			return nil, fmt.Errorf("missing aud `%s` in claim", aud)
		}
	}

	if err = claim.Validate(jwt.Expected{
		Subject: expSubj,
		Time:    time.Now().UTC(),
	}); err != nil {
		return nil, convertError(err)
	}

	return claim, convertError(err)
}
