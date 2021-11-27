package rest

import (
	"fmt"
	"net/http"
	"unicode/utf8"

	"prisma/tms/log"

	restful "github.com/orolia/go-restful"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/microcosm-cc/bluemonday"
)

// RFC 5988 Web Linking
func AddPaginationHeaderSafely(request *restful.Request, response *restful.Response, pagination *PaginationQuery) {
	if pagination != nil {
		if pagination.AfterId == "" {
			previousSkip := pagination.Skip - pagination.Limit
			linkValue := ""
			if previousSkip > -1 {
				linkValue += fmt.Sprintf("<%s?limit=%d&skip=%d>; rel=\"previous\"", request.SelectedRoutePath(), pagination.Limit, previousSkip)
			}
			nextSkip := pagination.Skip + pagination.Count
			linkValue += fmt.Sprintf(",<%s?limit=%d&skip=%d>; rel=\"next\"", request.SelectedRoutePath(), pagination.Limit, nextSkip)
			response.AddHeader("Link", linkValue)
		} else {
			linkValue := fmt.Sprintf("<%s?limit=%d&before=%s>; rel=\"previous\"", request.SelectedRoutePath(), pagination.Limit, pagination.BeforeId)
			linkValue += fmt.Sprintf(",<%s?limit=%d&after=%s>; rel=\"next\"", request.SelectedRoutePath(), pagination.Limit, pagination.AfterId)
			response.AddHeader("Link", linkValue)
		}
	}
}

func OutputEncodingAndSanitize(input string) string {
	if !utf8.ValidString(input) {
		return fmt.Sprint([]byte(input))
	}

	p := bluemonday.StrictPolicy()
	return p.Sanitize(input)
}

func WriteHeaderAndEntitySafely(response *restful.Response, status int, value interface{}) error {
	// TODO check
	return response.WriteHeaderAndEntity(status, value)
}

func WriteEntitySafely(response *restful.Response, value interface{}) error {
	// TODO check
	return response.WriteEntity(value)
}

func WriteValidationErrsSafely(response *restful.Response, errs []ErrorValidation) error {
	// TODO check
	return response.WriteHeaderAndEntity(http.StatusBadRequest, errs)
}

func WriteHeaderAndProtoSafely(response *restful.Response, status int, value proto.Message) error {
	// TODO check
	response.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	response.WriteHeader(status)
	marshaller := jsonpb.Marshaler{}
	if value != nil {
		return marshaller.Marshal(response.ResponseWriter, value)
	}
	return nil
}

func WriteProtoSpliceSafely(response *restful.Response, values []proto.Message) error {
	response.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	marshaller := jsonpb.Marshaler{}
	var err error
	response.ResponseWriter.Write([]byte("["))
	length := len(values)
	for index, value := range values {
		err = marshaller.Marshal(response.ResponseWriter, value)
		if index < length-1 {
			response.ResponseWriter.Write([]byte(","))
		}
	}
	if err == nil {
		response.ResponseWriter.Write([]byte("]"))
	} else {
		log.Error(err.Error(), err)
		response.WriteHeader(http.StatusInternalServerError)
	}
	return err
}

func WriteProtoSafely(response *restful.Response, value proto.Message) error {
	// TODO check
	response.Header().Set(restful.HEADER_ContentType, restful.MIME_JSON)
	marshaller := jsonpb.Marshaler{}
	return marshaller.Marshal(response.ResponseWriter, value)
}
