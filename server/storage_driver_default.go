//go:build !stor_bdb && !stor_sql
// +build !stor_bdb,!stor_sql

package server

import _ "github.com/mandiant/gocrack/server/storage/bdb"
