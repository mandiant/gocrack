package web

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/fireeye/gocrack/server/authentication"
	"github.com/fireeye/gocrack/server/storage"

	"github.com/gin-gonic/gin"
)

const badPassword = "The password you entered does not meet the requirements. It must have a length greater than or equal to 8, contain at least 1 special character, and 1 number."

// UserDetailedItem returns all information about a user, minus sensitive information. This should mimick storage.User
type UserDetailedItem struct {
	UserUUID     string    `json:"user_uuid"`
	Username     string    `json:"username"`
	Password     string    `json:"-"`
	Enabled      *bool     `json:"enabled,omitempty"`
	EmailAddress string    `json:"email_address"`
	IsSuperUser  bool      `json:"is_admin"`
	CreatedAt    time.Time `json:"created_at"`
}

// UserListingItem is an item returned in a user listing API and should mimick storage.User
type UserListingItem struct {
	UserUUID  string    `json:"user_uuid"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

// EditUserRequest must match `storage.UserModifyRequest`
type EditUserRequest struct {
	CurrentPassword *string `json:"current_password"`
	Password        *string `json:"new_password"`
	UserIsAdmin     *bool   `json:"user_is_admin"`
	Email           *string `json:"email"`
}

// EditUserResponse is returned on a successful update
type EditUserResponse struct {
	Modified bool `json:"modified"`
	EditUserRequest
}

// CreateUserRequest describes the payload sent by the user to create a new user
type CreateUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

// Validate returns any errors that are associated with the user input of CreateUserRequest
func (s CreateUserRequest) Validate() error {
	if s.Username == "" || len(s.Username) <= 4 {
		return errors.New("Username must be at least 4 characters or more")
	}

	// Super dumb email validation to check for the existence of an @ and period. Email validation is tough these days!
	if s.Email == "" || !strings.ContainsAny(s.Email, "@ & .") {
		return errors.New("Email must not be a valid email address")
	}

	if solidPassword := authentication.CheckPasswordRequirement(s.Password); !solidPassword {
		return errors.New(badPassword)
	}

	return nil
}

func (s *Server) webGetUsers(c *gin.Context) *WebAPIError {
	users, err := s.stor.GetUsers()
	if err != nil {
		if err == storage.ErrNotFound {
			c.JSON(http.StatusOK, []UserListingItem{})
			return nil
		}
		return &WebAPIError{
			StatusCode: http.StatusInternalServerError,
			Err:        err,
			UserError:  "The server was unable to process your request. Please try again later",
		}
	}

	resp := make([]UserListingItem, len(users))
	for i, user := range users {
		resp[i] = UserListingItem{
			Username:  user.Username,
			UserUUID:  user.UserUUID,
			CreatedAt: user.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, resp)
	return nil
}

func (s *Server) webEditUser(c *gin.Context) *WebAPIError {
	var (
		req    EditUserRequest
		userid = c.Param("user_uuid")
	)
	claim := getClaimInformation(c)

	if err := c.BindJSON(&req); err != nil {
		return &WebAPIError{
			StatusCode: http.StatusBadRequest,
			Err:        err,
			UserError:  "Invalid JSON",
		}
	}

	userrec, err := s.stor.GetUserByID(userid)
	if err != nil {
		if err == storage.ErrNotFound {
			return &WebAPIError{
				StatusCode: http.StatusNotFound,
				UserError:  "User not found or you do not have permission to view this record",
			}
		}
		return &WebAPIError{
			StatusCode: http.StatusInternalServerError,
			Err:        err,
		}
	}

	// The request to modify a user does not match their own record and the requester is not an admin, deny
	if userrec.UserUUID != claim.UserUUID && !claim.IsAdmin {
		return &WebAPIError{
			StatusCode: http.StatusNotFound,
			Err:        fmt.Errorf("%s attempted to modify %s's user record and did not have the proper rights", claim.UserUUID, userrec.UserUUID),
			UserError:  "User not found or you do not have permission to view this record",
		}
	}

	// User is unable to modify password because the record
	if (req.Password != nil && !s.auth.UserCanChangePassword()) || (req.CurrentPassword == nil && !claim.IsAdmin) {
		req.Password = nil
	} else if req.Password != nil && s.auth.UserCanChangePassword() {
		// Validate that the current password is correct if they're not an administrator
		if !claim.IsAdmin {
			if _, err := s.auth.Login(claim.Username, *req.CurrentPassword, claim.APIOnly); err != nil {
				return &WebAPIError{
					StatusCode: http.StatusBadRequest,
					Err:        err,
					UserError:  "Your current password is incorrect",
				}
			}
		}
		// Generate the secure password to set in our database
		securePassword, err := s.auth.GenerateSecurePassword(*req.Password)
		if err != nil {
			if err == authentication.ErrFailsRequirements || err == authentication.ErrPasswordEmpty {
				return &WebAPIError{
					StatusCode: http.StatusBadRequest,
					UserError:  badPassword,
				}
			}
			return &WebAPIError{
				StatusCode: http.StatusInternalServerError,
				Err:        err,
				UserError:  "We were unable to edit the user",
			}
		}
		req.Password = &securePassword
	}

	// User is unable to modify their admin flag if they arent an admin!
	if req.UserIsAdmin != nil && !claim.IsAdmin {
		req.UserIsAdmin = nil
	}

	if err = s.stor.EditUser(userid, storage.UserModifyRequest{
		Password:    req.Password,
		Email:       req.Email,
		UserIsAdmin: req.UserIsAdmin,
	}); err != nil {
		return &WebAPIError{
			StatusCode: http.StatusInternalServerError,
			Err:        err,
			UserError:  "We were unable to edit the user",
		}
	}

	c.JSON(http.StatusOK, &EditUserResponse{
		Modified:        true,
		EditUserRequest: req,
	})
	return nil
}

func (s *Server) webGetUser(c *gin.Context) *WebAPIError {
	var userid = c.Param("user_uuid")
	claim := getClaimInformation(c)

	userrec, err := s.stor.GetUserByID(userid)
	if err != nil {
		if err == storage.ErrNotFound {
			return &WebAPIError{
				StatusCode: http.StatusNotFound,
				UserError:  "User not found or you do not have permission to view this record",
			}
		}

		return &WebAPIError{
			StatusCode: http.StatusInternalServerError,
			Err:        err,
		}
	}

	// The request to modify a user does not match their own record and the requester is not an admin, deny
	if userrec.UserUUID != claim.UserUUID && !claim.IsAdmin {
		return &WebAPIError{
			StatusCode: http.StatusNotFound,
			Err:        fmt.Errorf("%s attempted to view %s's user record and did not have the proper rights", claim.UserUUID, userrec.UserUUID),
			UserError:  "User not found or you do not have permission to view this record",
		}
	}

	c.JSON(http.StatusOK, UserDetailedItem(*userrec))
	return nil
}

func (s *Server) webRegisterNewUser(c *gin.Context) *WebAPIError {
	var req CreateUserRequest

	if !s.auth.CanUsersRegister() {
		return &WebAPIError{
			StatusCode: http.StatusBadRequest,
			UserError:  "Registration is not allowed. Please contact your site administrator for more details",
		}
	}

	if err := c.BindJSON(&req); err != nil {
		return &WebAPIError{
			StatusCode: http.StatusBadRequest,
			Err:        err,
			UserError:  "Invalid JSON",
		}
	}

	if err := req.Validate(); err != nil {
		return &WebAPIError{
			StatusCode: http.StatusBadRequest,
			Err:        err,
			CanErrorBeShownToUser: true,
		}
	}

	if err := s.auth.CreateUser(storage.User{
		Username:     req.Username,
		EmailAddress: req.Email,
		Password:     req.Password,
	}); err != nil {
		if err == storage.ErrAlreadyExists {
			return &WebAPIError{
				StatusCode: http.StatusBadRequest,
				UserError:  "The username you picked is not available",
			}
		}
		return &WebAPIError{
			StatusCode: http.StatusInternalServerError,
			Err:        err,
			UserError:  "We were unable to create the user",
		}
	}

	c.Status(http.StatusCreated)
	return nil
}
