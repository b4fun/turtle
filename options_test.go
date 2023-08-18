package turtle

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Target_defaults(t *testing.T) {
	v := &Target{}
	assert.NoError(t, v.defaults())
}