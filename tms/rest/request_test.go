package rest

import (
	"net/http"
	"strings"
	"testing"

	restful "github.com/orolia/go-restful" 
	"github.com/go-openapi/spec"
	"github.com/stretchr/testify/assert"
)

var (
	testSchema = spec.Schema{
		SchemaProps: spec.SchemaProps{
			Required: []string{"Prop1"},
			Properties: map[string]spec.Schema{
				"Prop1": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{2}[0],
						MaxLength: &[]int64{255}[0],
						Pattern:   "[a-z0-9]",
					},
					SwaggerSchemaProps: spec.SwaggerSchemaProps{
						ReadOnly: true,
					},
				},
				"prop2": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{2}[0],
						MaxLength: &[]int64{255}[0],
					},
				},
			},
		},
	}
	testSchemaByTag = spec.Schema{
		SchemaProps: spec.SchemaProps{
			Required: []string{"prop2", "prop3"},
			Properties: map[string]spec.Schema{
				"prop4": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{1}[0],
						MaxLength: &[]int64{32}[0],
						Pattern:   "[0-9]",
					},
				},
				"prop5": {
					SchemaProps: spec.SchemaProps{
						Enum: []interface{}{4, 3, 2, 1},
					},
				},
				"prop6": {
					SchemaProps: spec.SchemaProps{
						Enum: []interface{}{"a", "b", "c", "d"},
					},
				},
				"prop7.prop8": {
					SchemaProps: spec.SchemaProps{
						MinLength: &[]int64{1}[0],
						MaxLength: &[]int64{32}[0],
						Pattern:   "[a-z0-9]",
					},
				},
			},
		},
	}
	testPathParameter = spec.Parameter{
		ParamProps: spec.ParamProps{
			Name:     "test-id",
			In:       "path",
			Required: true,
			Schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					MinLength: &[]int64{24}[0],
					MaxLength: &[]int64{24}[0],
					Pattern:   "[0-9a-fA-F]{24}",
				},
			},
		},
		SimpleSchema: spec.SimpleSchema{
			Type:   "string",
			Format: "hexadecimal",
		},
	}
	testPathParameterUuid = spec.Parameter{
		ParamProps: spec.ParamProps{
			Name:     "test-id",
			In:       "path",
			Required: true,
			Schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Pattern: "[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}",
				},
			},
		},
		SimpleSchema: spec.SimpleSchema{
			Type:   "string",
			Format: "hexadecimal",
		},
	}
	testPathParameterNoSchema = spec.Parameter{
		ParamProps: spec.ParamProps{
			Name:     "test-id",
			In:       "path",
			Required: true,
		},
	}
	testCookie = spec.Parameter{
		ParamProps: spec.ParamProps{
			Name:     "id",
			In:       "cookie",
			Required: true,
			Schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					MinLength: &[]int64{36}[0],
					MaxLength: &[]int64{36}[0],
					Pattern:   "[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}",
				},
			},
		},
		SimpleSchema: spec.SimpleSchema{
			Type:   "string",
			Format: "hexadecimal",
		},
	}
	testCookieNoSchema = spec.Parameter{
		ParamProps: spec.ParamProps{
			Name:     "id",
			In:       "cookie",
			Required: true,
		},
	}
)

type testEmptyEntity struct{}

type testEntity struct {
	Prop1 string `json:"prop1"`
	Prop2 string `json:"prop2,omitempty"`
}

type testEntityWithTag struct {
	Prop2 string `json:"prop2"`
	Prop3 string `json:"prop3,omitempty"`
	Prop4 int32 `json:"prop4"`
	Prop5 int32 `json:"prop5"`
	Prop6 string `json:"prop6"`
	Prop7 *testEntityNested `json:"prop7"`
}

type testEntityNested struct {
	Prop8 string `json:"prop8"`
}

func TestSanitizeValidateCookieOk(t *testing.T) {
	httpRequest, _ := http.NewRequest("GET", "/test", nil)
	httpRequest.AddCookie(&http.Cookie{
		Name:     "id",
		Value:    "e572eb5e-668e-433e-9c98-ce09e9d4e874",
		Path:     "/",
		HttpOnly: true,
	})
	request := restful.NewRequest(httpRequest)
	value, errs := SanitizeValidateCookie(request, testCookie)
	assert.Nil(t, errs)
	assert.Equal(t, "e572eb5e-668e-433e-9c98-ce09e9d4e874", value)
}

func TestSanitizeValidateCookieInvalid(t *testing.T) {
	httpRequest, _ := http.NewRequest("GET", "/test", nil)
	httpRequest.AddCookie(&http.Cookie{
		Name:     "id",
		Value:    "595a5f2e63209e3c9e0c5b6e",
		Path:     "/",
		HttpOnly: true,
	})
	request := restful.NewRequest(httpRequest)
	value, errs := SanitizeValidateCookie(request, testCookie)
	assert.NotNil(t, errs)
	assert.Equal(t, "", value)
}
func TestSanitizeValidateCookieRequired(t *testing.T) {
	httpRequest, _ := http.NewRequest("GET", "/test", nil)
	httpRequest.AddCookie(&http.Cookie{
		Name:     "id",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
	})
	request := restful.NewRequest(httpRequest)
	_, errs := SanitizeValidateCookie(request, testCookie)
	assert.NotNil(t, errs)
	assert.Equal(t, "Required", errs[0].Rule)
}

func TestSanitizeValidateCookieNoSchema(t *testing.T) {
	httpRequest, _ := http.NewRequest("GET", "/test", nil)
	httpRequest.AddCookie(&http.Cookie{
		Name:     "id",
		Value:    "e572eb5e-668e-433e-9c98-ce09e9d4e874",
		Path:     "/",
		HttpOnly: true,
	})
	request := restful.NewRequest(httpRequest)
	value, errs := SanitizeValidateCookie(request, testCookieNoSchema)
	assert.Nil(t, errs)
	assert.Equal(t, "e572eb5e-668e-433e-9c98-ce09e9d4e874", value)
}

func TestSanitizeValidateCookieXSS(t *testing.T) {
	httpRequest, _ := http.NewRequest("GET", "/test", nil)
	httpRequest.AddCookie(&http.Cookie{
		Name:     "id",
		Value:    "e572eb5e <script>alert()</SCRIPT>",
		Path:     "/",
		HttpOnly: true,
	})
	request := restful.NewRequest(httpRequest)
	value, errs := SanitizeValidateCookie(request, testCookie)
	assert.NotNil(t, errs)
	assert.Equal(t, "", value)
}

func TestSanitizeValidateCookieMinLength(t *testing.T) {
	httpRequest, _ := http.NewRequest("GET", "/test", nil)
	httpRequest.AddCookie(&http.Cookie{
		Name:     "id",
		Value:    "e572eb5e",
		Path:     "/",
		HttpOnly: true,
	})
	request := restful.NewRequest(httpRequest)
	_, errs := SanitizeValidateCookie(request, testCookie)
	assert.NotNil(t, errs)
	assert.Equal(t, "MinLength", errs[0].Rule)
}

func TestSanitizeValidateCookieMaxLength(t *testing.T) {
	httpRequest, _ := http.NewRequest("GET", "/test", nil)
	httpRequest.AddCookie(&http.Cookie{
		Name:     "id",
		Value:    "e572eb5ee572eb5ee572eb5ee572eb5ee572eb5e",
		Path:     "/",
		HttpOnly: true,
	})
	request := restful.NewRequest(httpRequest)
	_, errs := SanitizeValidateCookie(request, testCookie)
	assert.NotNil(t, errs)
	assert.Equal(t, "MaxLength", errs[0].Rule)
}

func TestSanitizeValidateCookiePattern(t *testing.T) {
	httpRequest, _ := http.NewRequest("GET", "/test", nil)
	httpRequest.AddCookie(&http.Cookie{
		Name:     "id",
		Value:    "e572eb5e-668e-433e-9c98-ce09e9d4e8--",
		Path:     "/",
		HttpOnly: true,
	})
	request := restful.NewRequest(httpRequest)
	_, errs := SanitizeValidateCookie(request, testCookie)
	assert.NotNil(t, errs)
	assert.Equal(t, "Pattern", errs[0].Rule)
}

func TestSanitizeValidatePathParameter(t *testing.T) {
	httpRequest, _ := http.NewRequest("GET", "/test/595a5f2e63209e3c9e0c5b6e", nil)
	request := restful.NewRequest(httpRequest)
	request.PathParameters()["test-id"] = "595a5f2e63209e3c9e0c5b6e"
	value, errs := SanitizeValidatePathParameter(request, testPathParameter)
	assert.Nil(t, errs)
	assert.Equal(t, "595a5f2e63209e3c9e0c5b6e", value)
}

func TestPathParameter(t *testing.T) {
	httpRequest, _ := http.NewRequest("GET", "/test/595a5f2e63209e3c9e0c5b6e", nil)
	request := restful.NewRequest(httpRequest)
	request.PathParameters()["test-id"] = "595a5f2e63209e3c9e0c5b6e"
	value, errs := SanitizeValidatePathParameter(request, testPathParameterNoSchema)
	assert.Nil(t, errs)
	assert.Equal(t, "595a5f2e63209e3c9e0c5b6e", value)
}

func TestPathParameterRequired(t *testing.T) {
	httpRequest, _ := http.NewRequest("GET", "/test", nil)
	request := restful.NewRequest(httpRequest)
	_, errs := SanitizeValidatePathParameter(request, testPathParameter)
	assert.NotNil(t, errs)
	assert.Equal(t, "Required", errs[0].Rule)
}

func TestSanitizePathParameter(t *testing.T) {
	httpRequest, _ := http.NewRequest("GET", "/test/595a5f2e63209e3c9e0c5b6e<script>alert()</SCRIPT>", nil)
	request := restful.NewRequest(httpRequest)
	request.PathParameters()["test-id"] = "595a5f2e63209e3c9e0c5b6e<script>alert()</SCRIPT>"
	value, errs := SanitizeValidatePathParameter(request, testPathParameterNoSchema)
	assert.Nil(t, errs)
	assert.Equal(t, "595a5f2e63209e3c9e0c5b6e", value)
}

func TestValidatePathParameter(t *testing.T) {
	httpRequest, _ := http.NewRequest("GET", "/test/badlength", nil)
	request := restful.NewRequest(httpRequest)
	request.PathParameters()["test-id"] = "badlength"
	_, errs := SanitizeValidatePathParameter(request, testPathParameter)
	assert.NotNil(t, errs)
	assert.Equal(t, "MinLength", errs[0].Rule)
}

func TestValidatePathParameterUuid(t *testing.T) {
	httpRequest, _ := http.NewRequest("GET", "/test/fe8385c6-af50-454d-bcf5-ac5969642993", nil)
	request := restful.NewRequest(httpRequest)
	request.PathParameters()["test-id"] = "fe8385c6-af50-454d-bcf5-ac5969642993"
	value, errs := SanitizeValidatePathParameter(request, testPathParameterUuid)
	assert.Nil(t, errs)
	assert.Equal(t, "fe8385c6-af50-454d-bcf5-ac5969642993", value)
}

func TestValidatePathParameterMax(t *testing.T) {
	httpRequest, _ := http.NewRequest("GET", "/test/badlength595a5f2e63209e3c9e0c5b6e", nil)
	request := restful.NewRequest(httpRequest)
	request.PathParameters()["test-id"] = "badlength595a5f2e63209e3c9e0c5b6e"
	_, errs := SanitizeValidatePathParameter(request, testPathParameter)
	assert.NotNil(t, errs)
	assert.Equal(t, "MaxLength", errs[0].Rule)
}

func TestValidatePathParameterPattern(t *testing.T) {
	httpRequest, _ := http.NewRequest("GET", "/test/5--a5f2e63209e3c9e0c5b6e", nil)
	request := restful.NewRequest(httpRequest)
	request.PathParameters()["test-id"] = "5--a5f2e63209e3c9e0c5b6e"
	_, errs := SanitizeValidatePathParameter(request, testPathParameter)
	assert.NotNil(t, errs)
	assert.Equal(t, "Pattern", errs[0].Rule)
}

func TestValidatePathParameterPatternHex(t *testing.T) {
	httpRequest, _ := http.NewRequest("GET", "/test/5--a5f2e63209e3c9e0c5b6e", nil)
	request := restful.NewRequest(httpRequest)
	request.PathParameters()["test-id"] = "5zza5f2e63209e3c9e0c5b6e"
	_, errs := SanitizeValidatePathParameter(request, testPathParameter)
	assert.NotNil(t, errs)
	assert.Equal(t, "Pattern", errs[0].Rule)
}

func TestSanitizeValidateReadEntity(t *testing.T) {
	bodyReader := strings.NewReader(`{"Prop1" : "42"}`)
	httpRequest, _ := http.NewRequest("GET", "/test", bodyReader)
	httpRequest.Header.Set("Content-Type", "application/json")
	request := &restful.Request{Request: httpRequest}
	tEntity := new(testEntity)
	err := SanitizeValidateReadEntity(request, testSchema, tEntity)
	assert.Nil(t, err)
	assert.Equal(t, "42", tEntity.Prop1)
}

func TestValidateRequired(t *testing.T) {
	bodyReader := strings.NewReader(`{}`)
	httpRequest, _ := http.NewRequest("GET", "/test", bodyReader)
	httpRequest.Header.Set("Content-Type", "application/json")
	request := &restful.Request{Request: httpRequest}
	tEntity := new(testEntity)
	errors := SanitizeValidateReadEntity(request, testSchema, tEntity)
	assert.NotNil(t, errors)
}

func TestValidateOptionalWithValueMinLength(t *testing.T) {
	bodyReader := strings.NewReader(`{"prop1":"abc","prop2":"2"}`)
	httpRequest, _ := http.NewRequest("GET", "/test", bodyReader)
	httpRequest.Header.Set("Content-Type", "application/json")
	request := &restful.Request{Request: httpRequest}
	tEntity := new(testEntity)
	errors := SanitizeValidateReadEntity(request, testSchema, tEntity)
	assert.NotNil(t, errors)
	assert.Equal(t, 1, len(errors))
	assert.Equal(t, "2", tEntity.Prop2)
}

func TestValidateOptionalWithoutValueMinLength(t *testing.T) {
	bodyReader := strings.NewReader(`{"prop1":"abc"}`)
	httpRequest, _ := http.NewRequest("GET", "/test", bodyReader)
	httpRequest.Header.Set("Content-Type", "application/json")
	request := &restful.Request{Request: httpRequest}
	tEntity := new(testEntity)
	errors := SanitizeValidateReadEntity(request, testSchema, tEntity)
	assert.Nil(t, errors)
	assert.Equal(t, "abc", tEntity.Prop1)
}

func TestValidateRequiredByTag(t *testing.T) {
	bodyReader := strings.NewReader(`{"prop2":"42","prop3":"24","prop4":12,"prop5":1,"prop6":"a"}`)
	httpRequest, _ := http.NewRequest("GET", "/test", bodyReader)
	httpRequest.Header.Set("Content-Type", "application/json")
	request := &restful.Request{Request: httpRequest}
	tEntity := new(testEntityWithTag)
	errors := SanitizeValidateReadEntity(request, testSchemaByTag, tEntity)
	assert.Nil(t, errors)
	assert.Equal(t, "42", tEntity.Prop2)
	assert.Equal(t, "24", tEntity.Prop3)
	assert.EqualValues(t, 12, tEntity.Prop4)
	assert.EqualValues(t, 1, tEntity.Prop5)
	assert.Equal(t, "a", tEntity.Prop6)
}

func TestValidateNested(t *testing.T) {
	bodyReader := strings.NewReader(`{"prop2":"42","prop3":"24","prop4":12,"prop5":1,"prop6":"a","prop7": {"prop8": "abc"}}`)
	httpRequest, _ := http.NewRequest("GET", "/test", bodyReader)
	httpRequest.Header.Set("Content-Type", "application/json")
	request := &restful.Request{Request: httpRequest}
	tEntity := new(testEntityWithTag)
	errors := SanitizeValidateReadEntity(request, testSchemaByTag, tEntity)
	assert.Nil(t, errors)
	assert.Equal(t, "42", tEntity.Prop2)
	assert.Equal(t, "24", tEntity.Prop3)
	assert.EqualValues(t, 12, tEntity.Prop4)
	assert.EqualValues(t, 1, tEntity.Prop5)
	assert.Equal(t, "a", tEntity.Prop6)
	assert.EqualValues(t, "abc", tEntity.Prop7.Prop8)
}

func TestValidateRequiredEmpty(t *testing.T) {
	bodyReader := strings.NewReader(`{}`)
	httpRequest, _ := http.NewRequest("GET", "/test", bodyReader)
	httpRequest.Header.Set("Content-Type", "application/json")
	request := &restful.Request{Request: httpRequest}
	tEntity := new(testEmptyEntity)
	errors := SanitizeValidateReadEntity(request, testSchema, tEntity)
	assert.NotNil(t, errors)
}

func TestSanitize1(t *testing.T) {
	bodyReader := strings.NewReader(`{"Prop1" : "'';!--"<XSS>=&{()}"}`)
	httpRequest, _ := http.NewRequest("GET", "/test", bodyReader)
	httpRequest.Header.Set("Content-Type", "application/json")
	request := &restful.Request{Request: httpRequest}
	tEntity := new(testEntity)
	err := SanitizeValidateReadEntity(request, testSchema, tEntity)
	assert.NotNil(t, err)
	assert.Equal(t, "", tEntity.Prop1)
}

func TestSanitize2(t *testing.T) {
	bodyReader := strings.NewReader(`{"Prop1" : "<SCRIPT SRC=http://xss.rocks/xss.js></SCRIPT>"}`)
	httpRequest, _ := http.NewRequest("GET", "/test", bodyReader)
	httpRequest.Header.Set("Content-Type", "application/json")
	request := &restful.Request{Request: httpRequest}
	tEntity := new(testEntity)
	err := SanitizeValidateReadEntity(request, testSchema, tEntity)
	assert.Nil(t, err)
	assert.Equal(t, "", tEntity.Prop1)
}

// JSON sanitize tests
// https://github.com/OWASP/json-sanitizer/blob/master/src/test/java/com/google/json/JsonSanitizerTest.java

// Script fragments
func TestJavaScriptFragments(t *testing.T) {
	bodyReader := strings.NewReader(`{"Prop1" : "42<script>alert()</SCRIPT>"}`)
	httpRequest, _ := http.NewRequest("GET", "/test", bodyReader)
	httpRequest.Header.Set("Content-Type", "application/json")
	request := &restful.Request{Request: httpRequest}
	tEntity := new(testEntity)
	err := SanitizeValidateReadEntity(request, testSchema, tEntity)
	assert.Nil(t, err)
	assert.Equal(t, "42", tEntity.Prop1)
}

// src='...'
func TestJavaScriptSrc(t *testing.T) {
	bodyReader := strings.NewReader(`{"Prop1" : "<script src=http://xss.rocks/xss.js></script>42"}`)
	httpRequest, _ := http.NewRequest("GET", "/test", bodyReader)
	httpRequest.Header.Set("Content-Type", "application/json")
	request := &restful.Request{Request: httpRequest}
	tEntity := new(testEntity)
	err := SanitizeValidateReadEntity(request, testSchema, tEntity)
	assert.Nil(t, err)
	assert.Equal(t, "42", tEntity.Prop1)
}

// lonely script tags
func TestLonelyScriptTags(t *testing.T) {
	bodyReader := strings.NewReader(`{"Prop1" : "</script>42"}`)
	httpRequest, _ := http.NewRequest("GET", "/test", bodyReader)
	httpRequest.Header.Set("Content-Type", "application/json")
	request := &restful.Request{Request: httpRequest}
	tEntity := new(testEntity)
	err := SanitizeValidateReadEntity(request, testSchema, tEntity)
	assert.Nil(t, err)
	assert.Equal(t, "42", tEntity.Prop1)
}
