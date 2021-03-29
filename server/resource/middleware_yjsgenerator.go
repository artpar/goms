package resource

import (
	"encoding/base64"
	"fmt"
	"github.com/artpar/api2go"
	"github.com/artpar/ydb"
	"github.com/buraksezer/olric"
	log "github.com/sirupsen/logrus"
	"strings"
)

type yjsHandlerMiddleware struct {
	dtopicMap        *map[string]*olric.DTopic
	cruds            *map[string]*DbResource
	documentProvider ydb.DocumentProvider
}

func (pc yjsHandlerMiddleware) String() string {
	return "EventGenerator"
}

func NewYJSHandlerMiddleware(documentProvider ydb.DocumentProvider) DatabaseRequestInterceptor {
	return &yjsHandlerMiddleware{
		dtopicMap:        nil,
		cruds:            nil,
		documentProvider: documentProvider,
	}
}

func (pc *yjsHandlerMiddleware) InterceptAfter(dr *DbResource, req *api2go.Request, results []map[string]interface{}) ([]map[string]interface{}, error) {

	return results, nil

}

func (pc *yjsHandlerMiddleware) InterceptBefore(dr *DbResource, req *api2go.Request, objects []map[string]interface{}) ([]map[string]interface{}, error) {

	requestMethod := strings.ToLower(req.PlainRequest.Method)
	switch requestMethod {
	case "get":
		break
	case "post":
		break

	case "update":
		fallthrough
	case "patch":

		for _, obj := range objects {
			reference_id := ""
			if requestMethod != "post" {
				reference_id = obj["reference_id"].(string)

			}

			for _, column := range dr.TableInfo().Columns {
				if BeginsWith(column.ColumnType, "file.") {
					fileColumnValue, ok := obj[column.ColumnName]
					if !ok {
						continue
					}
					fileColumnValueArray := fileColumnValue.([]interface{})

					existingYjsDocument := false
					if len(fileColumnValueArray) > 1 {
						existingYjsDocument = true
					}

					for i, fileInterface := range fileColumnValueArray {

						file := fileInterface.(map[string]interface{})

						if file["type"] == "x-crdt/yjs" {
							continue
						}

						var documentName = fmt.Sprintf("%v.%v.%v", dr.tableInfo.TableName, reference_id, column.ColumnName)
						document := pc.documentProvider.GetDocument(ydb.YjsRoomName(documentName))
						if document != nil {
							var documentHistory []byte
							documentHistory = document.GetInitialContentBytes()

							if !existingYjsDocument {
								fileColumnValueArray = append(fileColumnValueArray, map[string]interface{}{
									"contents": "x-crdt/yjs," + base64.StdEncoding.EncodeToString(documentHistory),
									"name":     file["name"].(string) + ".yjs",
									"type":     "x-crdt/yjs",
									"path":     file["path"],
								})

							} else {
								// yes remember the trick ?
								fileColumnValueArray[1-i] = map[string]interface{}{
									"contents": "x-crdt/yjs," + base64.StdEncoding.EncodeToString(documentHistory),
									"name":     file["name"].(string) + ".yjs",
									"type":     "x-crdt/yjs",
									"path":     file["path"],
								}
							}

							obj[column.ColumnName] = fileColumnValueArray
						}

					}
				}
			}
		}

		break
	case "delete":

		break
	default:
		log.Errorf("Invalid method: %v", req.PlainRequest.Method)
	}

	//currentUserId := context.Get(req.PlainRequest, "user_id").(string)
	//currentUserGroupId := context.Get(req.PlainRequest, "usergroup_id").([]string)

	return objects, nil

}