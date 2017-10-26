package authentication

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckPasswordRequirement(t *testing.T) {
	for _, test := range []struct {
		Password string
		IsStrong bool
	}{
		{
			Password: "!totallynotvalid",
			IsStrong: false,
		},
		{
			Password: "!Totally1AValidPassword",
			IsStrong: true,
		},
		{
			Password: "nolong",
			IsStrong: false,
		},
		{
			Password: "1Приветмир!",
			IsStrong: true,
		},
	} {
		isStrong := CheckPasswordRequirement(test.Password)
		assert.Equal(t, test.IsStrong, isStrong)
	}
}

func BenchCheckPasswordRequirement(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CheckPasswordRequirement("!@Tot2lly1AValidPassword")
	}
}
