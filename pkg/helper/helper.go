package helper

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
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

func ReplaceQueryParams(namedQuery string, params map[string]interface{}) (string, []interface{}) {
	var (
		i    int = 1
		args []interface{}
	)

	for k, v := range params {
		if k != "" && strings.Contains(namedQuery, ":"+k) {
			namedQuery = strings.ReplaceAll(namedQuery, ":"+k, "$"+strconv.Itoa(i))
			args = append(args, v)
			i++
		}
	}

	return namedQuery, args
}

func TableVer(ctx context.Context, conn *pgxpool.Pool, id, slug string) (map[string]interface{}, error) {

	query := `SELECT 
			"id",
			"slug",
			"label",
			"description",
			"show_in_menu",
			"subtitle_field_slug",
			"is_cached",
			"with_increment_id",
			"soft_delete",
			"digit_number"
	 FROM "table" WHERE `

	value := id

	var (
		label             string
		description       string
		showInMenu        bool
		subtitleFieldSlug string
		isCached          bool
		withIncrementId   bool
		softDelete        bool
		digitNumber       int32
	)

	if id != "" {
		query += ` "id" = $1`
	} else if slug != "" {
		query += ` "slug" = $1`
		value = slug
	}

	err := conn.QueryRow(ctx, query, value).Scan(&id, &slug, &label)
	if err != nil {
		return map[string]interface{}{}, err
	}

	return map[string]interface{}{
		"id":                  id,
		"slug":                slug,
		"label":               label,
		"description":         description,
		"show_in_menu":        showInMenu,
		"subtitle_field_slug": subtitleFieldSlug,
		"is_cached":           isCached,
		"with_increment_id":   withIncrementId,
		"soft_delete":         softDelete,
		"digit_number":        digitNumber,
	}, nil

}

func BoardOrderChecker(ctx context.Context, conn *pgxpool.Pool, table_slug string) error {

	var table_id string

	query := `SELECT id FROM "table" WHERE slug = $1`

	err := conn.QueryRow(ctx, query, table_slug).Scan(&table_id)
	if err != nil {
		fmt.Println("error while getting id from table: BoardOrderChecker", err)
		return err
	}

	var boardOrderID string

	query2 := `SELECT id FROM "field" WHERE table_id = $1 AND "slug" = 'board_order'`
	err = conn.QueryRow(ctx, query2, table_id).Scan(&boardOrderID)
	if err == pgx.ErrNoRows {
		now := time.Now()
		query3 := `INSERT INTO field (id, table_id, required, slug, label, "default", "type", "index", attributes, is_visible, autofill_field, autofill_table, created_at, updated_at)
						  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`

		_, err = conn.Exec(ctx, query3, "93999892-78b0-4674-9e42-6a2a41524ebe",
			table_id, false, "board_order", "BOARD ORDER", "", "NUMBER",
			"string", "{'fields': {'icon': {'stringValue': '', 'kind': 'stringValue'}, 'placeholder': {'stringValue': '', 'kind': 'stringValue'}, 'showTooltip': {'boolValue': false, 'kind': 'boolValue'}}",
			false, "", "", now, now)
		if err != nil {
			fmt.Println("Error While inserting data into BoardOrderChecker", err)
			return err
		}
	} else if err != nil {
		return err
	}

	return nil
}
