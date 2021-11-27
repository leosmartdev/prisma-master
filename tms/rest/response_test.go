package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	restful "github.com/orolia/go-restful" 
	"github.com/stretchr/testify/assert"
)

func TestResponseWriteHeaderAndEntity(t *testing.T) {
	prop1 := OutputEncodingAndSanitize(`Hello, 世界<script>xss</script>`)
	tEntity := &testEntity{Prop1: prop1}
	rec := httptest.NewRecorder()
	rsp := restful.NewResponse(rec)
	rsp.SetRequestAccepts(restful.MIME_JSON)
	rsp.PrettyPrint(false)
	err := rsp.WriteHeaderAndEntity(http.StatusOK, tEntity)
	assert.Nil(t, err)
	assert.Equal(t, `{"prop1":"Hello, 世界"}
`, rec.Body.String())
}

func TestResponseWriteHeaderAndEntity2(t *testing.T) {
	prop1 := OutputEncodingAndSanitize(string([]byte{0xff, 0xfe, 0xfd}))
	tEntity := &testEntity{Prop1: prop1}
	rec := httptest.NewRecorder()
	rsp := restful.NewResponse(rec)
	rsp.SetRequestAccepts(restful.MIME_JSON)
	rsp.PrettyPrint(false)
	err := rsp.WriteHeaderAndEntity(http.StatusOK, tEntity)
	assert.Nil(t, err)
	assert.Equal(t, `{"prop1":"[255 254 253]"}
`, rec.Body.String())
}
