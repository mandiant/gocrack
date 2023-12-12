//go:build !auth_database && !auth_ldap
// +build !auth_database,!auth_ldap

package server

import (
	_ "github.com/mandiant/gocrack/server/authentication/database"
	_ "github.com/mandiant/gocrack/server/authentication/ldap"
)
