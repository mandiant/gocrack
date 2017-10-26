// +build !auth_database,!auth_ldap

package server

import (
	_ "github.com/fireeye/gocrack/server/authentication/database"
	_ "github.com/fireeye/gocrack/server/authentication/ldap"
)
