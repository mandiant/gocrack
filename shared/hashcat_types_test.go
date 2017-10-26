package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLookupHashcatHashType(t *testing.T) {
	ht := LookupHashcatHashType(0)
	if ht == nil {
		assert.Fail(t, "ht lookup failed")
		return
	}

	assert.Equal(t, "MD5", ht.Name)

	ht = LookupHashcatHashType(1333333337)
	assert.Nil(t, ht)
}
