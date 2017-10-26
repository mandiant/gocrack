package shared

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestHumanDuration(t *testing.T) {
	hd := HumanDuration{time.Minute * 5}
	b, err := yaml.Marshal(&hd)
	assert.Nil(t, err)
	assert.Equal(t, "5m0s\n", string(b))

	var new HumanDuration
	err = yaml.Unmarshal(b, &new)
	assert.Nil(t, err)
	assert.Equal(t, int64(300000000000), new.Nanoseconds())

	err = yaml.Unmarshal([]byte("notValid"), &new)
	assert.NotNil(t, err)
}
