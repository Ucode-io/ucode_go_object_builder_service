package helper

import (
	"context"
	"encoding/json"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

// func ReplaceQueryParams(namedQuery string, params map[string]interface{}) (string, []interface{}) {
// 	var (
// 		i    int = 1
// 		args []interface{}
// 	)

// 	for k, v := range params {
// 		if k != "" {
// 			oldsize := len(namedQuery)
// 			namedQuery = strings.ReplaceAll(namedQuery, ":"+k, "$"+strconv.Itoa(i))

// 			if oldsize != len(namedQuery) {
// 				args = append(args, v)
// 				i++
// 			}
// 		}
// 	}

// 	return namedQuery, args
// }

func ChangeHostname(data []byte) ([]byte, error) {

	var (
		isChangedByHost = map[string]bool{}
	)

	if err := json.Unmarshal(data, &isChangedByHost); err != nil {
		return nil, err
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	isChangedByHost[hostname] = true

	data, err = json.Marshal(isChangedByHost)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// func TableVersion(conn pgxpool.Pool)

func GetTableByIdSlug(ctx context.Context, conn *pgxpool.Pool, id, slug string) (map[string]interface{}, error) {

	query := `SELECT id, slug, label FROM "table" WHERE `

	value := id

	var label string

	if id != "" {
		query += ` id = $1`
	} else if slug != "" {
		query += ` slug = $1`
		value = slug
	}

	err := conn.QueryRow(ctx, query, value).Scan(&id, &slug, &label)
	if err != nil {
		return map[string]interface{}{}, err
	}

	return map[string]interface{}{
		"id":    id,
		"slug":  slug,
		"label": label,
	}, nil
}

func GetFieldBySlug(ctx context.Context, conn *pgxpool.Pool, slug string, tableId string) (map[string]interface{}, error) {

	query := `SELECT id, type, attributes FROM "field" WHERE slug = $1 AND table_id = $2`

	var (
		id, ftype  string
		attributes []byte
	)

	err := conn.QueryRow(ctx, query, slug, tableId).Scan(&id, &slug)
	if err != nil {
		return map[string]interface{}{}, err
	}

	return map[string]interface{}{
		"id":         id,
		"type":       ftype,
		"attributes": attributes,
	}, nil
}
