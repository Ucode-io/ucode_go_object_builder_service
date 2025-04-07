package formula

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"ucode/ucode_go_object_builder_service/models"
	"ucode/ucode_go_object_builder_service/pkg/helper"
	psqlpool "ucode/ucode_go_object_builder_service/pool"

	"github.com/pkg/errors"
	"github.com/spf13/cast"
)

func CalculateFormulaBackend(ctx context.Context, conn *psqlpool.Pool, attributes map[string]any, tableSlug string) (map[string]float32, error) {
	var (
		query         string
		response      = make(map[string]float32)
		relationField = tableSlug + "_id"
		table         = strings.Split(cast.ToString(attributes["table_from"]), "#")[0]
		field         = cast.ToString(attributes["sum_field"])
		round         = cast.ToInt(attributes["number_of_rounds"])
		formulaFilter []FormulaFilter
	)

	if attributes["formula_filters"] != nil {
		formulaFilterByte, err := json.Marshal(attributes["formula_filters"])
		if err != nil {
			return map[string]float32{}, errors.Wrap(err, "CalculateFormulaBackend - json.Marshal")
		}

		err = json.Unmarshal(formulaFilterByte, &formulaFilter)
		if err != nil {
			return map[string]float32{}, errors.Wrap(err, "CalculateFormulaBackend - json.Unmarshal")
		}
	}

	whereClause, params, err := buildWhereClause(formulaFilter)

	switch cast.ToString(attributes["type"]) {
	case "SUMM":
		query = fmt.Sprintf(`SELECT %s, SUM(%s) FROM  "%s" %s GROUP BY %s`, relationField, field, table, whereClause, relationField)
	case "MAX":
		query = fmt.Sprintf(`SELECT %s, MAX(%s) FROM "%s" %s GROUP BY %s`, relationField, field, table, whereClause, relationField)
	case "AVG":
		query = fmt.Sprintf(`SELECT %s, AVG(%s) FROM "%s" %s GROUP BY %s`, relationField, field, table, whereClause, relationField)
	}

	rows, err := conn.Query(ctx, query, params...)
	if err != nil {
		return map[string]float32{}, errors.Wrap(err, "CalculateFormulaBackend - conn.Query")
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id     string
			num    float32
			scanId sql.NullString
		)

		err = rows.Scan(&scanId, &num)
		if err != nil {
			return map[string]float32{}, errors.Wrap(err, "CalculateFormulaBackend - rows.Scan")
		}

		id = scanId.String

		if round > 0 {
			format := "%." + fmt.Sprint(round) + "f"
			num = cast.ToFloat32(fmt.Sprintf(format, num))
		}
		response[id] = num
	}

	return response, nil
}

func CalculateFormulaFrontend(attributes map[string]any, fields []models.Field, object map[string]any) (any, error) {
	computedFormula := attributes["formula"].(string)

	for _, el := range fields {
		value, ok := object[el.Slug]
		if !ok {
			value = 0
		}

		if floatValue, ok := value.(float64); ok {
			valueStr := fmt.Sprintf("%f", floatValue)
			computedFormula = strings.ReplaceAll(computedFormula, el.Slug, valueStr)
			continue
		}

		valueStr := fmt.Sprintf("%v", value)
		computedFormula = strings.ReplaceAll(computedFormula, el.Slug, valueStr)
	}

	result, err := helper.CallJS(computedFormula)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func buildWhereClause(filters []FormulaFilter) (string, []any, error) {
	var (
		whereClauses []string
		params       []any
	)

	for i, filter := range filters {
		keyParts := strings.Split(filter.Key, "#")
		if len(keyParts) == 0 {
			return "", nil, fmt.Errorf("invalid filter key at index %d", i)
		}
		field := keyParts[0]

		switch v := filter.Value.(type) {
		case []any:
			// Handle array values with IN clause
			if len(v) == 0 {
				continue
			}
			placeholders := make([]string, len(v))
			for i := range v {
				placeholders[i] = fmt.Sprintf("$%d", len(params)+1+i)
			}
			params = append(params, v...)
			whereClauses = append(whereClauses, fmt.Sprintf("%s IN (%s)", field, strings.Join(placeholders, ",")))
		default:
			// Handle single value
			params = append(params, v)
			whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", field, len(params)))
		}
	}

	return fmt.Sprintf(" WHERE %s", strings.Join(whereClauses, " AND ")), params, nil
}

type FormulaFilter struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
}
