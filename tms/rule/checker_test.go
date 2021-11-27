package rule

import (
	"testing"
	"github.com/json-iterator/go/assert"
)

func TestCompRegExValue(t *testing.T) {
	ret, err := compRegExValue("[a-z]+", "abcdjkf")
	assert.NoError(t, err)
	assert.True(t, ret)

	ret, err = compRegExValue("[a-z]+$", "eqe21312121")
	assert.NoError(t, err)
	assert.False(t, ret)
}
