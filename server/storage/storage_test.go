package storage

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStorageError(t *testing.T) {
	customError := StorageError{
		DriverError: errors.New("get to the chopper!"),
	}
	assert.Equal(t, "get to the chopper!", customError.Error())
}
