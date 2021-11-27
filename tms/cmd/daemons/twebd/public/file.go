package public

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"prisma/gogroup"
	"prisma/tms/db/mongo"
	"prisma/tms/log"
	"prisma/tms/moc"
	"prisma/tms/rest"

	"prisma/tms/security"

	restful "github.com/orolia/go-restful" 
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/go-openapi/spec"
)

const (
	MaxMemorySize = 10 * 1024 * 1024
	FILE_CLASSID  = "File"
)

var (
	parameterFileId = spec.Parameter{
		ParamProps: spec.ParamProps{
			Name:     "file-id",
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
)

type FileRest struct {
	client *mongo.MongoClient
	group  gogroup.GoGroup
}

type IdResponse struct {
	Id string `json:"id"`
}

func NewFileRest(client *mongo.MongoClient, group gogroup.GoGroup) *FileRest {
	return &FileRest{
		client: client,
		group:  group,
	}
}

// curl -X PUT -F 'file=@./hello_world.txt' http://localhost:8080/api/v2/file
func (f *FileRest) Create(request *restful.Request, response *restful.Response) {
	ACTION := moc.File_CREATE.String()
	if !security.HasPermissionForAction(request.Request.Context(), FILE_CLASSID, ACTION) {
		security.Audit(request.Request.Context(), FILE_CLASSID, ACTION, "FAIL")
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
		return
	}

	err := request.Request.ParseMultipartForm(MaxMemorySize)
	if err != nil {
		security.Audit(request.Request.Context(), FILE_CLASSID, ACTION, "FAIL_ERROR")
		response.WriteError(http.StatusInternalServerError,
			fmt.Errorf("unable to parse multipart form: %v", err))
		return
	}

	src, header, err := request.Request.FormFile("file")
	if err != nil {
		security.Audit(request.Request.Context(), FILE_CLASSID, ACTION, "FAIL_ERROR")
		response.WriteError(http.StatusInternalServerError,
			fmt.Errorf("unable to find file in form: %v", err))
		return
	}
	defer src.Close()
	contentType := header.Header.Get("content-type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	db := f.client.DB()
	defer f.client.Release(db)
	dest, err := db.GridFS("fs").Create(header.Filename)
	if err != nil {
		security.Audit(request.Request.Context(), FILE_CLASSID, ACTION, "FAIL_ERROR")
		response.WriteError(http.StatusInternalServerError,
			fmt.Errorf("unable to create file in store: %v", err))
		return
	}
	defer dest.Close()
	dest.SetContentType(contentType)

	_, err = io.Copy(dest, src)
	if err != nil {
		security.Audit(request.Request.Context(), FILE_CLASSID, ACTION, "FAIL_ERROR")
		response.WriteError(http.StatusInternalServerError,
			fmt.Errorf("unable to copy file to store: %v", err))
		return
	}
	id, ok := dest.Id().(bson.ObjectId)
	if !ok {
		security.Audit(request.Request.Context(), FILE_CLASSID, ACTION, "FAIL_ERROR")
		response.WriteError(http.StatusInternalServerError,
			errors.New("expecting object id"))
		return
	}

	security.Audit(request.Request.Context(), FILE_CLASSID, ACTION, "SUCCESS")
	response.WriteHeaderAndEntity(http.StatusOK, IdResponse{
		Id: fmt.Sprintf("%v", id.Hex()),
	})
}

func (f *FileRest) Get(request *restful.Request, response *restful.Response) {
	ACTION := moc.File_READ.String()
	if !security.HasPermissionForAction(request.Request.Context(), FILE_CLASSID, ACTION) {
		security.Audit(request.Request.Context(), FILE_CLASSID, ACTION, "FAIL")
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
		return
	}
	fileHexID, errs := rest.SanitizeValidatePathParameter(request, parameterFileId)
	if !valid(errs, request, response, FILE_CLASSID, ACTION) {
		return
	}
	fileID := bson.ObjectIdHex(fileHexID)
	db := f.client.DB()
	defer f.client.Release(db)
	var src *mgo.GridFile
	src, err := db.GridFS("fs").OpenId(fileID)
	if err != nil {
		// try meta id
		err = db.GridFS("fs").Find(bson.M{"metadata.id": fileHexID}).One(&src)
		if err == nil {
			src, err = db.GridFS("fs").OpenId(src.Id())
		}
		if err != nil {
			security.Audit(request.Request.Context(), FILE_CLASSID, ACTION, "FAIL")
			response.WriteErrorString(http.StatusNotFound, "not found")
			return
		}
	}
	defer src.Close()

	disposition := fmt.Sprintf("filename=\"%v\"", src.Name())
	if src.ContentType() == "application/octet-stream" {
		disposition = "attachment; " + disposition
	}

	response.AddHeader("Content-Type", src.ContentType())
	response.AddHeader("Content-Disposition", disposition)
	response.AddHeader("Content-Length", strconv.FormatInt(src.Size(), 10))
	response.WriteHeader(http.StatusOK)

	_, err = io.Copy(response, src)
	if err != nil {
		log.Debug("download terminated by remote host")
	}
}

func (f *FileRest) Delete(request *restful.Request, response *restful.Response) {
	ACTION := moc.File_DELETE.String()
	if !security.HasPermissionForAction(request.Request.Context(), FILE_CLASSID, ACTION) {
		security.Audit(request.Request.Context(), FILE_CLASSID, ACTION, "FAIL")
		response.WriteError(http.StatusForbidden, security.ErrorForbidden)
		return
	}

	fileHexID := request.PathParameter("file-id")
	if !bson.IsObjectIdHex(fileHexID) {
		security.Audit(request.Request.Context(), FILE_CLASSID, ACTION, "FAIL")
		response.WriteErrorString(http.StatusBadRequest, "invalid id")
		return
	}

	fileID := bson.ObjectIdHex(fileHexID)
	db := f.client.DB()
	defer f.client.Release(db)
	if err := db.GridFS("fs").RemoveId(fileID); err != nil {
		security.Audit(request.Request.Context(), FILE_CLASSID, ACTION, "FAIL")
		response.WriteErrorString(http.StatusNotFound, "not found")
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, IdResponse{
		Id: fmt.Sprintf("%v", fileHexID),
	})
}
