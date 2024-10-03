package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"ucode/ucode_go_object_builder_service/config"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/models"
	"ucode/ucode_go_object_builder_service/pkg/helper"
	psqlpool "ucode/ucode_go_object_builder_service/pool"
	"ucode/ucode_go_object_builder_service/storage"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"github.com/tealeg/xlsx"
	"github.com/xuri/excelize/v2"
)

type excelRepo struct {
	db *pgxpool.Pool
}

func NewExcelRepo(db *pgxpool.Pool) storage.ExcelRepoI {
	return &excelRepo{
		db: db,
	}
}

func (e *excelRepo) ExcelRead(ctx context.Context, req *nb.ExcelReadRequest) (resp *nb.ExcelReadResponse, err error) {
	cfg := config.Load()

	endpoint := cfg.MinioHost
	accessKeyID := cfg.MinioAccessKeyID
	secretAccessKey := cfg.MinioSecretKey
	fileID := req.Id

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: true,
	})
	if err != nil {
		return &nb.ExcelReadResponse{}, err
	}

	bucketName := "docs"
	fileObjectKey := fileID + ".xlsx"
	createFilePath := "./" + fileObjectKey

	err = downloadFile(minioClient, bucketName, fileObjectKey, createFilePath)
	if err != nil {
		return &nb.ExcelReadResponse{}, err
	}

	objectRow, err := readFirstRow(createFilePath)
	if err != nil {
		return &nb.ExcelReadResponse{}, err
	}

	err = os.Remove(createFilePath)
	if err != nil {
		return &nb.ExcelReadResponse{}, err
	}

	return &nb.ExcelReadResponse{Rows: objectRow}, nil
}

func (e *excelRepo) ExcelToDb(ctx context.Context, req *nb.ExcelToDbRequest) (resp *nb.ExcelToDbResponse, err error) {
	var (
		conn            = psqlpool.Get(req.GetProjectId())
		cfg             = config.Load()
		endpoint        = cfg.MinioHost
		accessKeyID     = cfg.MinioAccessKeyID
		secretAccessKey = cfg.MinioSecretKey
		i               int
		fieldsMap       = make(map[string]models.Field)
		slugsMap        = make(map[string]string)
	)

	tx, err := conn.Begin(ctx)
	if err != nil {
		return &nb.ExcelToDbResponse{}, errors.Wrap(err, "conn.Begin")
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: true,
	})
	if err != nil {
		return &nb.ExcelToDbResponse{}, errors.Wrap(err, "minio.New")
	}

	object, err := minioClient.GetObject(context.Background(), "docs", req.Id+".xlsx", minio.GetObjectOptions{})
	if err != nil {
		return &nb.ExcelToDbResponse{}, errors.Wrap(err, "minioClient.GetObject")
	}
	defer object.Close()

	localFileName := "localfile.xlsx"
	localFile, err := os.Create(localFileName)
	if err != nil {
		return &nb.ExcelToDbResponse{}, errors.Wrap(err, "os.Create")
	}
	defer localFile.Close()

	if _, err := io.Copy(localFile, object); err != nil {
		return &nb.ExcelToDbResponse{}, errors.Wrap(err, "io.Copy")
	}

	f, err := excelize.OpenFile(localFileName)
	if err != nil {
		return &nb.ExcelToDbResponse{}, errors.Wrap(err, "excelize.OpenFile")
	}

	query := `SELECT f.id, f.slug, f.type FROM "field" f JOIN "table" t ON f.table_id = t.id WHERE t.slug = $1`

	fieldRows, err := tx.Query(ctx, query, req.TableSlug)
	if err != nil {
		return &nb.ExcelToDbResponse{}, errors.Wrap(err, "tx.Query")
	}
	defer fieldRows.Close()

	data, err := helper.ConvertStructToMap(req.Data)
	if err != nil {
		return &nb.ExcelToDbResponse{}, errors.Wrap(err, "ConvertStructToMap")
	}

	fields := []models.Field{}

	for fieldRows.Next() {
		var (
			id, slug, ftype string
		)

		err = fieldRows.Scan(
			&id,
			&slug,
			&ftype,
		)
		if err != nil {
			return &nb.ExcelToDbResponse{}, errors.Wrap(err, "fieldRows.Scan")
		}

		fieldsMap[id] = models.Field{
			Id:   id,
			Slug: slug,
			Type: ftype,
		}
		fieldsMap[slug] = models.Field{
			Id:   id,
			Slug: slug,
			Type: ftype,
		}
		fields = append(fields, models.Field{
			Id:   id,
			Slug: slug,
			Type: ftype,
		})
	}

	for {
		cell, err := f.GetCellValue(sh, letters[i]+"1")
		if err != nil {
			return &nb.ExcelToDbResponse{}, errors.Wrap(err, "GetCellValue")
		}
		if cell == "" {
			break
		}

		fieldId := cast.ToString(data[cell])
		slug := fieldsMap[fieldId]

		slugsMap[letters[i]] = slug.Slug
		i++
	}

	rows, err := f.GetRows(sh)
	if err != nil {
		return &nb.ExcelToDbResponse{}, errors.Wrap(err, "GetRows")
	}

	var fullData = []map[string]interface{}{}

	for c, row := range rows {
		if c == 0 {
			continue
		}

		body := make(map[string]interface{})

		for i, cell := range row {
			var value interface{}
			if cell != "" {
				field := fieldsMap[slugsMap[letters[i]]]

				if helper.FIELD_TYPES[field.Type] == "FLOAT" {
					value = cast.ToInt(cell)
				} else if field.Type == "MULTISELECT" {
					value = strings.Split(cell, ",")
				} else if field.Type == "SWITCH" || field.Type == "CHECKBOX" {
					if strings.ToUpper(cell) == "ИСТИНА" || strings.ToUpper(cell) == "TRUE" {
						value = true
					} else {
						value = false
					}
				} else {
					value = cell
				}

				body[slugsMap[letters[i]]] = value
			}
		}
		fullData = append(fullData, body)
	}

	query, args, err := MakeQueryForMultiInsert(ctx, tx, req.TableSlug, fullData, fields)
	if err != nil {
		return &nb.ExcelToDbResponse{}, errors.Wrap(err, "MakeQueryForMultiInsert")
	}

	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		return &nb.ExcelToDbResponse{}, errors.Wrap(err, "tx.Exec")
	}

	err = tx.Commit(ctx)
	if err != nil {
		return &nb.ExcelToDbResponse{}, errors.Wrap(err, "tx.Commit")
	}

	return &nb.ExcelToDbResponse{}, nil
}

func MakeQueryForMultiInsert(ctx context.Context, tx pgx.Tx, tableSlug string, data []map[string]interface{}, fields []models.Field) (string, []interface{}, error) {
	var (
		args       []interface{}
		argCount   = 1
		tableSlugs = []string{}
		fieldM     = make(map[string]helper.FieldBody)
		newFields  = []models.Field{}
		query      = fmt.Sprintf(`INSERT INTO %s (guid`, tableSlug)
	)

	for _, field := range fields {
		if field.Slug == "guid" {
			continue
		} else if field.Type == "INCREMENT_NUMBER" {
			continue
		}

		query += fmt.Sprintf(", %s", field.Slug)
	}

	query += ") VALUES"

	fQuery := ` SELECT
		f."id",
		f."type",
		f."attributes",
		f."relation_id",
		f."autofill_table",
		f."autofill_field",
		f."slug"
	FROM "field" f JOIN "table" as t ON f.table_id = t.id WHERE t.slug = $1`

	fieldRows, err := tx.Query(ctx, fQuery, tableSlug)
	if err != nil {
		return "", nil, err
	}
	defer fieldRows.Close()

	for fieldRows.Next() {
		field := models.Field{}

		var (
			atr           = []byte{}
			autoFillTable sql.NullString
			autoFillField sql.NullString
			relationId    sql.NullString
			attributes    = make(map[string]interface{})
		)

		err = fieldRows.Scan(
			&field.Id,
			&field.Type,
			&atr,
			&relationId,
			&autoFillTable,
			&autoFillField,
			&field.Slug,
		)
		if err != nil {
			return "", nil, err
		}

		if err := json.Unmarshal(atr, &field.Attributes); err != nil {
			return "", nil, err
		}
		if err := json.Unmarshal(atr, &attributes); err != nil {
			return "", nil, err
		}

		tableSlugs = append(tableSlugs, field.Slug)

		if _, ok := Ftype[field.Type]; ok {
			fieldM[field.Type] = helper.FieldBody{
				Slug:       field.Slug,
				Attributes: attributes,
			}
		}

		newFields = append(newFields, field)
	}

	reqBody := helper.CreateBody{
		FieldMap:   fieldM,
		Fields:     newFields,
		TableSlugs: tableSlugs,
	}

	for _, body := range data {

		structBody, err := helper.ConvertMapToStruct(body)
		if err != nil {
			return "", nil, err
		}

		body, _, err = helper.PrepareToCreateInObjectBuilderWithTx(ctx, tx, &nb.CommonMessage{
			Data:      structBody,
			TableSlug: tableSlug,
		}, reqBody)
		if err != nil {
			return "", nil, err
		}

		query += " ("
		for _, field := range fields {
			if field.Type == "INCREMENT_NUMBER" {
				continue
			}

			if field.Slug == "guid" {
				query += fmt.Sprintf(" $%d,", argCount)
				value, ok := body["guid"]
				if !ok || helper.IsEmpty(value) {
					args = append(args, uuid.NewString())
				} else {
					args = append(args, value)
				}
				argCount++
				continue
			}

			query += fmt.Sprintf(" $%d,", argCount)
			args = append(args, body[field.Slug])
			argCount++
		}

		query = strings.TrimRight(query, ",")
		query += "),"
	}

	query = strings.TrimRight(query, ",")

	return query, args, nil
}

func downloadFile(minioClient *minio.Client, bucketName, fileObjectKey, createFilePath string) error {
	object, err := minioClient.GetObject(context.Background(), bucketName, fileObjectKey, minio.GetObjectOptions{})
	if err != nil {
		return err
	}
	defer object.Close()

	file, err := os.Create(createFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, object)
	if err != nil {
		return fmt.Errorf("error copying object to file: %v", err)
	}

	return nil
}

func readFirstRow(filePath string) ([]string, error) {
	xlFile, err := xlsx.OpenFile(filePath)
	if err != nil {
		return nil, err
	}

	var firstRow []string
	if len(xlFile.Sheets) > 0 {
		firstSheet := xlFile.Sheets[0]
		if len(firstSheet.Rows) > 0 {
			for _, cell := range firstSheet.Rows[0].Cells {
				firstRow = append(firstRow, cell.String())
			}
		}
	}

	return firstRow, nil
}
