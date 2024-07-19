package postgres

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
	"ucode/ucode_go_object_builder_service/config"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/models"
	"ucode/ucode_go_object_builder_service/pkg/helper"
	psqlpool "ucode/ucode_go_object_builder_service/pkg/pool"
	"ucode/ucode_go_object_builder_service/storage"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/spf13/cast"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

type csvRepo struct {
	db *pgxpool.Pool
}

func NewCSVRepo(db *pgxpool.Pool) storage.CSVRepoI {
	return &csvRepo{
		db: db,
	}
}

func (o *csvRepo) GetListInCSV(ctx context.Context, req *nb.CommonMessage) (resp *nb.CommonMessage, err error) {

	conn := psqlpool.Get(req.GetProjectId())

	var (
		params = make(map[string]interface{})
	)

	paramBody, err := json.Marshal(req.Data)
	if err != nil {
		return &nb.CommonMessage{}, err
	}
	if err := json.Unmarshal(paramBody, &params); err != nil {
		return &nb.CommonMessage{}, err
	}

	fieldIds := cast.ToStringSlice(params["field_ids"])

	delete(params, "field_ids")

	query := `SELECT f.type, f.slug, f.attributes, f.label FROM "field" f WHERE f.id = ANY ($1)`

	fieldRows, err := conn.Query(ctx, query, pq.Array(fieldIds))
	if err != nil {
		return &nb.CommonMessage{}, err
	}
	defer fieldRows.Close()

	fields := make(map[string]models.Field)
	fieldsArr := []models.Field{}

	for fieldRows.Next() {
		var (
			fBody = models.Field{}
			attrb = []byte{}
		)

		err = fieldRows.Scan(
			&fBody.Type,
			&fBody.Slug,
			&attrb,
			&fBody.Label,
		)
		if err != nil {
			return &nb.CommonMessage{}, err
		}

		if err := json.Unmarshal(attrb, &fBody.Attributes); err != nil {
			return &nb.CommonMessage{}, err
		}

		fields[fBody.Slug] = fBody
		fieldsArr = append(fieldsArr, fBody)
	}

	items, _, err := helper.GetItems(ctx, conn, models.GetItemsBody{
		TableSlug: req.TableSlug,
		Params:    params,
		FieldsMap: fields,
	})
	if err != nil {
		return &nb.CommonMessage{}, err
	}

	filename := fmt.Sprintf("report_%d.csv", time.Now().Unix())
	filepath := "./" + filename
	file, err := os.Create(filename)
	if err != nil {
		return &nb.CommonMessage{}, err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := []string{}
	for _, field := range fieldsArr {
		headers = append(headers, field.Label)
	}
	if err := writer.Write(headers); err != nil {
		return &nb.CommonMessage{}, err
	}

	for _, item := range items {
		record := []string{}
		for _, field := range fieldsArr {
			if field.Type == "MULTI_LINE" {
				re := regexp.MustCompile(`<[^>]+>`)
				result := re.ReplaceAllString(cast.ToString(item[field.Slug]), "")
				item[field.Slug] = result
			} else if field.Type == "DATE" {
				timeF, err := time.Parse("2006-01-02", strings.Split(cast.ToString(item[field.Slug]), " ")[0])
				if err != nil {
					return &nb.CommonMessage{}, err
				}
				item[field.Slug] = timeF.Format("02.01.2006")
			} else if field.Type == "DATE_TIME" {
				newTime := strings.Split(cast.ToString(item[field.Slug]), " ")[0] + " " + strings.Split(cast.ToString(item[field.Slug]), " ")[1]
				timeF, err := time.Parse("2006-01-02 15:04:05", newTime)
				if err != nil {
					return &nb.CommonMessage{}, err
				}
				item[field.Slug] = timeF.Format("02.01.2006 15:04")
			} else if field.Type == "MULTISELECT" {
				attributes, err := helper.ConvertStructToMap(field.Attributes)
				if err != nil {
					return &nb.CommonMessage{}, err
				}
				multiselectValue := ""
				if options, ok := attributes["options"].([]interface{}); ok {
					values := cast.ToStringSlice(item[field.Slug])
					for _, val := range values {
						for _, op := range options {
							opt := cast.ToStringMap(op)
							if val == cast.ToString(opt["value"]) {
								if label, ok := opt["label"].(string); ok && label != "" {
									multiselectValue += label + ","
								} else {
									multiselectValue += cast.ToString(opt["value"]) + ","
								}
							}
						}
					}
				}
				item[field.Slug] = strings.TrimRight(multiselectValue, ",")
			}
			record = append(record, cast.ToString(item[field.Slug]))
		}
		if err := writer.Write(record); err != nil {
			return &nb.CommonMessage{}, err
		}
	}

	cfg := config.Load()

	endpoint := cfg.MinioHost
	accessKeyID := cfg.MinioAccessKeyID
	secretAccessKey := cfg.MinioSecretKey

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: true,
	})
	if err != nil {
		return &nb.CommonMessage{}, err
	}

	_, err = minioClient.FPutObject(
		context.Background(),
		"reports",
		filename,
		filepath,
		minio.PutObjectOptions{ContentType: "text/csv"},
	)
	if err != nil {
		return &nb.CommonMessage{}, err
	}

	err = os.Remove(filepath)
	if err != nil {
		return &nb.CommonMessage{}, err
	}

	link := fmt.Sprintf("%s/reports/%s", endpoint, filename)
	respCSV := map[string]string{
		"link": link,
	}

	marshaledInputMap, err := json.Marshal(respCSV)
	outputStruct := &structpb.Struct{}
	if err != nil {
		return &nb.CommonMessage{}, err
	}

	err = protojson.Unmarshal(marshaledInputMap, outputStruct)
	if err != nil {
		return &nb.CommonMessage{}, err
	}

	return &nb.CommonMessage{TableSlug: req.TableSlug, Data: outputStruct}, nil
}
