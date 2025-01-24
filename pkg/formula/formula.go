package formula

import (
	"context"
	"database/sql"
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
	)

	switch cast.ToString(attributes["type"]) {
	case "SUMM":
		query = fmt.Sprintf(`SELECT %s, SUM(%s) FROM "%s" GROUP BY %s`, relationField, field, table, relationField)
	case "MAX":
		query = fmt.Sprintf(`SELECT %s, MAX(%s) FROM "%s" GROUP BY %s`, relationField, field, table, relationField)
	case "AVG":
		query = fmt.Sprintf(`SELECT %s, AVG(%s) FROM "%s" GROUP BY %s`, relationField, field, table, relationField)
	}

	rows, err := conn.Query(ctx, query)
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
