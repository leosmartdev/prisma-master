package incident

import (
	"prisma/tms/test/context"

	"github.com/stretchr/testify/assert"

	"testing"
)

type testPrefixer struct {
	prefix string
}

func (p *testPrefixer) Prefix() string {
	return p.prefix
}

func TestCreateReferenceId(t *testing.T) {
	prefixer := testPrefixer{
		prefix: "Incident-",
	}
	refId := IdCreatorInstance(context.Test()).Next(&prefixer)
	assert.NotNil(t, refId, "not nil")
	assert.Contains(t, refId, "-")
	t.Log(refId)
}

func TestCreateReferenceIdTwice(t *testing.T) {
	prefixer := testPrefixer{
		prefix: "INC",
	}
	refId := IdCreatorInstance(context.Test()).Next(&prefixer)
	refId = IdCreatorInstance(context.Test()).Next(&prefixer)
	assert.NotNil(t, refId, "not nil")
	assert.NotEqual(t, "INC-1", refId)
	t.Log(refId)
}

func TestCreateReferenceIdFruit(t *testing.T) {
	prefixer := testPrefixer{
		prefix: "Fruit",
	}
	refId := IdCreatorInstance(context.Test()).Next(&prefixer)
	assert.NotNil(t, refId, "not nil")
	t.Log(refId)
}

func TestCreateReferenceIdEmpty(t *testing.T) {
	prefixer := testPrefixer{
		prefix: "",
	}
	refId := IdCreatorInstance(context.Test()).Next(&prefixer)
	assert.NotNil(t, refId, "not nil")
	assert.NotEqual(t, "", refId)
	assert.NotContains(t, refId, "-")
	t.Log(refId)
}

func TestCreateReferenceIdTwo(t *testing.T) {
	prefixer := testPrefixer{
		prefix: "Po-",
	}
	refId := IdCreatorInstance(context.Test()).Next(&prefixer)
	assert.NotNil(t, refId, "not nil")
	assert.Contains(t, refId, "-")
	t.Log(refId)
}

func TestCreateReferenceIdThree(t *testing.T) {
	prefixer := testPrefixer{
		prefix: "UFO-AB-",
	}
	refId := IdCreatorInstance(context.Test()).Next(&prefixer)
	assert.NotNil(t, refId, "not nil")
	assert.Contains(t, refId, "-")
	t.Log(refId)
}
