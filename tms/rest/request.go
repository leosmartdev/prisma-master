package rest

import (
	"context"
	"html"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"prisma/tms/security"

	restful "github.com/orolia/go-restful"
	"github.com/go-openapi/spec"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/microcosm-cc/bluemonday"
	"prisma/tms/moc"
)

const (
	mimeJsonProtobuf = "application/json-protobuf"
)

var (
	queryAnchor = spec.Parameter{
		ParamProps: spec.ParamProps{
			Name: "anchor",
			In:   "query",
			Schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					MinLength: &[]int64{1}[0],
					MaxLength: &[]int64{48}[0],
				},
			},
		},
	}
	queryBefore = spec.Parameter{
		ParamProps: spec.ParamProps{
			Name: "before",
			In:   "query",
			Schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					MinLength: &[]int64{1}[0],
					MaxLength: &[]int64{48}[0],
				},
			},
		},
	}
	queryAfter = spec.Parameter{
		ParamProps: spec.ParamProps{
			Name: "after",
			In:   "query",
			Schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					MinLength: &[]int64{1}[0],
					MaxLength: &[]int64{48}[0],
				},
			},
		},
	}
	querySort = spec.Parameter{
		ParamProps: spec.ParamProps{
			Name: "sort",
			In:   "query",
			Schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					MinLength: &[]int64{1}[0],
					MaxLength: &[]int64{48}[0],
				},
			},
		},
	}
)

// RouteMatcher checks the route and returns ClassId used in security functions
type RouteMatcher interface {
	// MatchRoute returns bool if there is a match and ClassId
	MatchRoute(route string) (bool, string)
}

type PaginationQuery struct {
	Limit    int
	Skip     int
	Count    int
	Sort     string // TODO change to map field,order - string,int
	AfterId  string
	BeforeId string
	Anchor   string
}

func SanitizePagination(request *restful.Request) (*PaginationQuery, bool) {
	anchor, _ := SanitizeValidateQueryParameter(request, queryAnchor)
	before, _ := SanitizeValidateQueryParameter(request, queryBefore)
	after, _ := SanitizeValidateQueryParameter(request, queryAfter)
	sort, _ := SanitizeValidateQueryParameter(request, querySort)
	limit, err := strconv.Atoi(request.QueryParameter("limit"))
	skip, _ := strconv.Atoi(request.QueryParameter("skip"))
	if err == nil {
		return &PaginationQuery{
			Limit:    limit,
			Skip:     skip,
			Sort:     sort,
			BeforeId: before,
			AfterId:  after,
			Anchor:   anchor,
		}, true
	}
	return nil, false
}

func SanitizeValidateCookie(req *restful.Request, p spec.Parameter) (value string, errs []ErrorValidation) {
	cookie, err := req.Request.Cookie(p.Name) // id
	if err != nil {
		errs = append(errs, ErrorValidation{
			Property: p.Name,
			Rule:     "Error",
			Message:  err.Error()})
		return
	}

	if cookie.Value == "" {
		errs = append(errs, ErrorValidation{
			Property: p.Name,
			Rule:     "Required",
			Message:  "Required non-empty property"})
	}

	if p.Schema != nil {
		currentLength := len(cookie.Value)
		if p.Schema.MinLength != nil && currentLength < int(*p.Schema.MinLength) {
			errs = append(errs, ErrorValidation{
				Property: p.Name,
				Rule:     "MinLength",
				Message:  "Invalid length"})
		}
		if p.Schema.MaxLength != nil && currentLength > int(*p.Schema.MaxLength) {
			errs = append(errs, ErrorValidation{
				Property: p.Name,
				Rule:     "MaxLength",
				Message:  "Invalid length"})
		}
		pattern := regexp.MustCompile(p.Schema.Pattern)
		if !pattern.MatchString(cookie.Value) {
			errs = append(errs, ErrorValidation{
				Property: p.Name,
				Rule:     "Pattern",
				Message:  "Invalid match " + p.Schema.Pattern})
		}
	}

	if len(errs) > 0 {
		return // whitelist -> reject
	}

	policy := bluemonday.StrictPolicy()
	value = policy.Sanitize(cookie.Value)
	return
}

// Authorized checks if the user in the context has permission to perform the action on the class id.
// Writes to audit log and response.
// Returns false if not authorized, and caller should not continue processing the transaction.
func Authorized(ctx context.Context, response *restful.Response, classId string, action string) bool {
	authorized := security.HasPermissionForAction(ctx, classId, action)
	if !authorized {
		security.Audit(ctx, classId, action, security.FAIL_PERMISSION)
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
	}
	return authorized
}

// Valid checks if validation errors are present.
// Writes to audit log and response.
// Returns false if validation errors, and caller should not continue processing the transaction.
func Valid(errs []ErrorValidation, ctx context.Context, response *restful.Response, classId string, action string) bool {
	valid := (len(errs) == 0)
	if !valid {
		security.Audit(ctx, classId, action, security.FAIL_VALIDATION)
		WriteValidationErrsSafely(response, errs)
	}
	return valid
}

func SanitizeValidatePathParameter(request *restful.Request, parameter spec.Parameter) (string, []ErrorValidation) {
	return sanitizeValidateParameter(request, parameter, request.PathParameter(parameter.Name))
}

func SanitizeValidateQueryParameter(request *restful.Request, parameter spec.Parameter) (string, []ErrorValidation) {
	return sanitizeValidateParameter(request, parameter, request.QueryParameter(parameter.Name))
}

func sanitizeValidateParameter(_ *restful.Request, parameter spec.Parameter, value string) (string, []ErrorValidation) {
	var errs []ErrorValidation
	if nil != parameter.Schema {
		valueLength := len(value)
		if "" == value {
			if parameter.Required {
				if "" == value {
					errs = append(errs, ErrorValidation{
						Property: parameter.Name,
						Rule:     "Required",
						Message:  "Required non-empty property"})
					return "", errs
				}
			}
			return "", nil
		}
		if parameter.Schema.MinLength != nil {
			if valueLength < int(*parameter.Schema.MinLength) {
				errs = append(errs, ErrorValidation{
					Property: parameter.Name,
					Rule:     "MinLength",
					Message:  "Invalid length"})
			}
		}
		if parameter.Schema.MaxLength != nil {
			if valueLength > int(*parameter.Schema.MaxLength) {
				errs = append(errs, ErrorValidation{
					Property: parameter.Name,
					Rule:     "MaxLength",
					Message:  "Invalid length"})
			}
		}
		pattern := regexp.MustCompile(parameter.Schema.Pattern)
		if !pattern.MatchString(value) {
			errs = append(errs, ErrorValidation{
				Property: parameter.Name,
				Rule:     "Pattern",
				Message:  "Invalid match " + parameter.Schema.Pattern})
		}
	}
	policy := bluemonday.StrictPolicy()
	value = policy.Sanitize(value)
	return value, errs
}

func SanitizeValidateReadProto(request *restful.Request, schema spec.Schema, entity interface{}) []ErrorValidation {
	request.Request.Header.Set(restful.HEADER_ContentType, mimeJsonProtobuf)
	return SanitizeValidateReadEntity(request, schema, entity)
}

// Sanitizes inputs (header, parameters, body).
// Validates input against Schema.
// Populates entity
// https://www.owasp.org/index.php/Input_Validation_Cheat_Sheet
// https://www.owasp.org/index.php/Data_Validation#Data_Validation_Strategies
// TODO convert to errors for localization https://github.com/epoberezkin/ajv-i18n/blob/master/messages/index.js
func SanitizeValidateReadEntity(request *restful.Request, schema spec.Schema, entity interface{}) []ErrorValidation {
	restful.RegisterEntityAccessor(mimeJsonProtobuf, entityJSONPROTOBUFAccess{})
	var errors []ErrorValidation
	// may require restful.RegisterEntityAccessor() call to registry new EntityReaderWriter
	// https://github.com/owasp/json-sanitizer
	// unmarshal
	err := request.ReadEntity(entity)
	if err != nil {
		errors = append(errors, ErrorValidation{
			Property: reflect.TypeOf(entity).Elem().String(),
			Rule:     "Unmarshal",
			Message:  err.Error(),
		})
		return errors
	}
	return SanitizeValidate(entity, schema)
}

func SanitizeValidate(entity interface{}, schema spec.Schema) []ErrorValidation {
	var errors []ErrorValidation
	// enables quick field lookup based on field name and json field tag
	fields := make(map[string]reflect.Value)
	typeEntity := reflect.TypeOf(entity).Elem()
	// Iterate over all available fields and read the tag value
	for i := 0; i < typeEntity.NumField(); i++ {
		// Get the field, returns https://golang.org/pkg/reflect/#StructField
		field := typeEntity.Field(i)
		// Get the field tag value
		tag := field.Tag.Get("json")
		if tag != "" {
			tag = strings.Split(tag, ",")[0]
			fields[tag] = reflect.Indirect(reflect.ValueOf(entity)).FieldByName(field.Name)
		}
		fields[field.Name] = reflect.Indirect(reflect.ValueOf(entity)).FieldByName(field.Name)
		// if nested then go deeper (just one level deeper)
		fieldValue := reflect.Indirect(reflect.ValueOf(entity)).FieldByName(field.Name)
		switch fieldValue.Kind() {
		case reflect.Ptr:
			fieldValueStruct := reflect.Indirect(fieldValue)
			if !(fieldValue.IsValid() && fieldValueStruct.IsValid()) {
				continue
			}

			fieldValuetypeEntity := reflect.TypeOf(fieldValue.Interface()).Elem()
			fieldIndex := fieldValueStruct.NumField() - 1
			for fieldIndex > -1 {
				// Get the field tag value
				fieldTag := fieldValuetypeEntity.Field(fieldIndex).Tag.Get("json")
				if fieldTag != "" {
					fieldTag = strings.Split(fieldTag, ",")[0]
					fields[tag+"."+fieldTag] = fieldValueStruct.Field(fieldIndex)
				}
				fieldKey := fieldValuetypeEntity.Field(fieldIndex).Name
				fields[tag+"."+fieldKey] = fieldValueStruct.Field(fieldIndex)
				fieldIndex -= 1
			}
		case reflect.Struct: // like a device.id
			if !fieldValue.IsValid() {
				continue
			}

			fieldValuetypeEntity := reflect.TypeOf(fieldValue.Interface())
			fieldIndex := fieldValue.NumField() - 1
			for fieldIndex > -1 {
				// Get the field tag value
				fieldTag := fieldValuetypeEntity.Field(fieldIndex).Tag.Get("json")
				if fieldTag != "" {
					fieldTag = strings.Split(fieldTag, ",")[0]
					fields[tag+"."+fieldTag] = fieldValue.Field(fieldIndex)
				}
				fieldKey := fieldValuetypeEntity.Field(fieldIndex).Name
				fields[tag+"."+fieldKey] = fieldValue.Field(fieldIndex)
				fieldIndex -= 1
			}
		}
	}
	// validate
	// check all property/fields are required
	for _, requiredField := range schema.Required {
		field, ok := fields[requiredField]
		if ok {
			if reflect.Invalid == field.Kind() {
				errors = append(errors, ErrorValidation{
					Property: requiredField,
					Rule:     "Required",
					Message:  "Required property"})
			} else if reflect.String == field.Kind() {
				// trim space CONV-1312
				str := strings.TrimSpace(field.String())
				if str == "" {
					errors = append(errors, ErrorValidation{
						Property: requiredField,
						Rule:     "Required",
						Message:  "Required non-empty property"})
				}
			} else if reflect.Slice == field.Kind() {
				if 0 == field.Len() {
					errors = append(errors, ErrorValidation{
						Property: requiredField,
						Rule:     "Required",
						Message:  "Required non-zero length"})
				}
			}
		} else {
			errors = append(errors, ErrorValidation{
				Property: requiredField,
				Rule:     "Required",
				Message:  "Required property"})
		}
	}
	// check schema properties.  whitelist
	for prop, whitelist := range schema.Properties {
		field, ok := fields[prop]
		if !ok {
			continue
		}
		str := field.String()
		// if empty then do not check
		if str != "" {
			// pattern
			pattern := regexp.MustCompile(whitelist.Pattern)
			if !pattern.MatchString(str) {
				errors = append(errors, ErrorValidation{
					Property: prop,
					Rule:     "Pattern",
					Message:  "Pattern mismatch"})
			}
			// max
			if whitelist.MaxLength != nil {
				if len(str) > int(*whitelist.MaxLength) {
					errors = append(errors, ErrorValidation{
						Property: prop,
						Rule:     "MaxLength",
						Message:  "Out of range"})
				}
			}
			// enum
			if whitelist.Enum != nil && field.Interface() != nil {
				inEnum := false
				for _, value := range whitelist.Enum {
					switch v := value.(type) {
					case string:
						inEnum = v == str
					case int:
						inEnum = int(v) == int(field.Int())
					case int32:
						inEnum = int32(v) == int32(field.Int())
					case int64:
						inEnum = int64(v) == int64(field.Int())
						//default:
						//	fmt.Println(v)
					}
					if inEnum {
						break
					}
				}
				if !inEnum {
					errors = append(errors, ErrorValidation{
						Property: prop,
						Rule:     "Enum",
						Message:  "Not in enum"})
				}
			}
		}
		// min
		if whitelist.MinLength != nil {
			// only check if not empty string, because optional, required check above will catch empty
			if len(str) > 0 && len(str) < int(*whitelist.MinLength) {
				errors = append(errors, ErrorValidation{
					Property: prop,
					Rule:     "MinLength",
					Message:  "Out of range."})
			}
		}
	}
	// sanitize  https://www.owasp.org/index.php/XSS_Filter_Evasion_Cheat_Sheet
	p := bluemonday.StrictPolicy()
	for prop := range schema.Properties {
		field, ok := fields[prop]
		if ok {
			if !field.CanSet() {
				continue
			}
			if reflect.String == field.Kind() {
				a := field.String()
				b := p.Sanitize(a)
				// allow unescaped strings in JSON CONV-1311
				field.SetString(html.UnescapeString(b))
			}
		}
	}
	return errors
}

// entityJSONAccess is a EntityReaderWriter for JSON encoding
type entityJSONPROTOBUFAccess struct {
	// This is used for setting the Content-Type header when writing
	ContentType string
}

// Read unmarshalls the value from JSON using jsonpb.
func (e entityJSONPROTOBUFAccess) Read(req *restful.Request, v interface{}) error {
	p, ok := v.(proto.Message)
	if ok {
		jspb, ok := p.(moc.JSPBUnmarshaler)
		if ok {
			return jspb.Unmarshal(req.Request.Body, p)
		} else {
			return jsonpb.Unmarshal(req.Request.Body, p)
		}
	}
	return nil
}

// Write marshalls the value to JSON and set the Content-Type Header using jsonpb.
func (e entityJSONPROTOBUFAccess) Write(resp *restful.Response, status int, v interface{}) error {
	if v == nil {
		resp.WriteHeader(status)
		return nil
	}
	p, ok := v.(proto.Message)
	if ok {
		resp.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
		resp.WriteHeader(status)
		marshaller := jsonpb.Marshaler{}
		return marshaller.Marshal(resp.ResponseWriter, p)
	}
	return nil
}
