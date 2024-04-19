package helper

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CheckRelationField struct {
	conn      *pgxpool.Pool
	FieldName string
	TableID   string
}

func CheckRelationFieldExists(ctx context.Context, req CheckRelationField) (bool, string, error) {

	rows, err := req.conn.Query(ctx, "SELECT slug FROM field WHERE table_id = $1 AND slug LIKE $2 ORDER BY slug DESC", req.TableID, req.FieldName+"%")
	if err != nil {
		return false, "", fmt.Errorf("failed to query fields: %v", err)
	}
	defer rows.Close()

	var lastField string
	for rows.Next() {
		var fieldSlug string
		err := rows.Scan(&fieldSlug)
		if err != nil {
			return false, "", fmt.Errorf("failed to scan field slug: %v", err)
		}
		lastField = fieldSlug
	}

	// If lastField is not empty, extract the index and increment it
	if lastField != "" {
		// Split the slug to extract the index
		parts := strings.Split(lastField, "_")
		if len(parts) > 1 {
			index, err := strconv.Atoi(parts[len(parts)-1])
			if err != nil {
				return false, "", fmt.Errorf("failed to parse index from last field: %v", err)
			}
			// Increment the index
			index++
			lastField = fmt.Sprintf("%s_%d", req.FieldName, index)
		}
	}

	// Return the existence status and the last field name
	return lastField != "", lastField, nil
}
