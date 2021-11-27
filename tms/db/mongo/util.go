package mongo

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"prisma/tms/db"
	"prisma/tms/rest"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

var (
	collectionMatcher = regexp.MustCompile(`collection:[^ ]+`)
)

func collectionName(q *mgo.Query) string {
	strval := fmt.Sprintf("%+v", q)
	return collectionMatcher.FindString(strval)
}

func getSession(ctx context.Context) (*mgo.Session, error) {
	//dialInfoValue := ctx.Value("mongodb")
	dialInfoValue := ctx.Value("mongodb")
	dialInfo, ok := dialInfoValue.(*mgo.DialInfo)
	if !ok {
		return nil, errors.New("Cast to *mgo.DialInfo failed. ctx: " + fmt.Sprint(ctx))
	}
	sess, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		return nil, err
	}
	if dialInfo.Mechanism == MongoDbX509Mechanism {
		credValue := ctx.Value("mongodb-cred")
		cred, ok := credValue.(*mgo.Credential)
		if !ok {
			return nil, errors.New("Cast to *mgo.Credential failed. ctx" + fmt.Sprint(ctx))
		}
		if sess.Login(cred) != nil {
			return nil, err
		}
	}
	return sess, err

}

func createMongoSort(sortFields db.SortFields) bson.D {
	sortElements := bson.D{}
	for _, fieldOrder := range sortFields {
		value := 1
		if fieldOrder.Desc {
			value = -1
		}
		sortElements = append(sortElements, bson.DocElem{
			Name:  fieldOrder.Field,
			Value: value,
		})
	}
	return sortElements
}

func createMongoQueryFromMap(searchMap map[string]string) bson.M {
	query := bson.M{}
	for k, v := range searchMap {
		values := strings.Split(v, ",")
		if len(values) > 1 {
			if strings.HasPrefix(values[0], "$") {
				if values[0] == "$exists" {
					existsValue, _ := strconv.ParseBool(values[1])
					query[k] = bson.M{values[0]: existsValue}
				} else {
					query[k] = bson.M{values[0]: values[1]}
				}
			} else {
				orQuery := make([]bson.M, 0)
				for _, value := range values {
					orQuery = append(orQuery, bson.M{k: value})
				}
				query["$or"] = orQuery
			}
		} else {
			query[k] = v
		}
	}
	return query
}

func createPipe(searchMap map[string]string, pagination *rest.PaginationQuery) []bson.M {
	query := createMongoQueryFromMap(searchMap)
	pipe := []bson.M{
		{"$match": query},
	}
	if pagination.Skip > 0 {
		pipe = append(pipe, bson.M{"$sort": bson.M{pagination.Sort: 1}})
		pipe = append(pipe, bson.M{"$skip": pagination.Skip})
	} else if pagination.AfterId != "" {
		query["_id"] = bson.M{"$gt": bson.ObjectIdHex(pagination.AfterId)}
		pipe = append(pipe, bson.M{"$sort": bson.M{pagination.Sort: 1}})
	} else if pagination.BeforeId != "" {
		query["_id"] = bson.M{"$lt": bson.ObjectIdHex(pagination.BeforeId)}
		pipe = append(pipe, bson.M{"$sort": bson.M{pagination.Sort: -1}})
	} else {
		pipe = append(pipe, bson.M{"$sort": bson.M{pagination.Sort: 1}})
	}
	if pagination.Anchor == "" {
		pipe = append(pipe, bson.M{"$limit": pagination.Limit})
	}
	return pipe
}
