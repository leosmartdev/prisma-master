package public

import (
	"strconv"
	"strings"

	"prisma/gogroup"
	"prisma/tms/rest"
	"prisma/tms/search"
	"prisma/tms/security"

	restful "github.com/orolia/go-restful" 
	"github.com/go-openapi/spec"
	"prisma/tms/log"
	"github.com/globalsign/mgo/bson"
	"net/http"
)

var (
	parameterQuery = spec.Parameter{
		ParamProps: spec.ParamProps{
			Name:     "query",
			In:       "query",
			Schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					MaxLength: &[]int64{2000}[0],
				},
			},
		},
	}
	parameterLimit = spec.Parameter{
		ParamProps: spec.ParamProps{
			Name:     "limit",
			In:       "query",
			Required: false,
			Schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					MaxLength: &[]int64{3}[0],
					Pattern:   "[0-9]{1,3}",
				},
			},
		},
	}
)

// SearchRest is a structure to determine rest api for search endpoints
type SearchRest struct {
	s search.Searcher
}

func NewSearchRest(_ gogroup.GoGroup, s search.Searcher) *SearchRest {
	return &SearchRest{
		s: s,
	}
}

// TODO: filter post body to avoid NOSQL-injection
func (r *SearchRest) SearchFunc(collectionName string) func(request *restful.Request, response *restful.Response) {
	return func(request *restful.Request, response *restful.Response) {
		const CLASSID = "Search"
		const ACTION = "READ"

		if !authorized(request, response, CLASSID, ACTION) {
			return
		}
		tables := make([]string, 0)
		if collectionName == "" {
			tablesString := request.PathParameter("tables")
			tables = strings.Split(tablesString, ",")
		} else {
			tables = append(tables, collectionName)
		}
		text, errs := rest.SanitizeValidateQueryParameter(request, parameterQuery)
		slimit, errs2 := rest.SanitizeValidateQueryParameter(request, parameterLimit)
		errs = append(errs, errs2...)
		if !valid(errs, request, response, CLASSID, ACTION) {
			return
		}
		if text == "" {
			text = request.PathParameter("query")
		}
		limit, _ := strconv.Atoi(slimit)
		if limit == 0 {
			limit = 1000
		}
		var queryFields bson.M
		// parse additional query fields from post body
		if request.Request.Method == http.MethodPost {
			if err := request.ReadEntity(&queryFields); !errorFree(err, request, response, CLASSID, ACTION) {
				return
			}
		}
		results, err := r.s.Search(text, queryFields, tables, limit)
		if !errorFree(err, request, response, CLASSID, ACTION) {
			return
		}

		security.Audit(request.Request.Context(), CLASSID, ACTION, "SUCCESS")
		if err := response.WriteEntity(results); err != nil {
			log.Error(err.Error())
		}
	}
}
