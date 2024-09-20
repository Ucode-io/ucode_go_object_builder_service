package helper

import (
	"context"
	"fmt"
	"ucode/ucode_go_object_builder_service/models"

	"github.com/spf13/cast"
)

func GetAdditional(ctx context.Context, req models.GetAdditionalRequest) ([]interface{}, error) {
	if additionalRequest, exist := req.Params["additional_request"]; exist {
		additionalRequestMap, ok := additionalRequest.(map[string]interface{})
		if ok {
			additionalValues := cast.ToStringSlice(additionalRequestMap["additional_values"])
			additionalField, ok := additionalRequestMap["additional_field"].(string)
			if ok && len(additionalValues) > 0 {
				var (
					filter    = fmt.Sprintf(" WHERE deleted_at IS NULL AND %s IN (", additionalField)
					resultMap = make(map[string]bool, len(req.Result))
					ids       []string
				)

				for _, obj := range req.Result {
					values := cast.ToStringMap(obj)
					resValue := cast.ToString(values[additionalField])
					resultMap[resValue] = true
				}

				for _, id := range additionalValues {
					if _, exist := resultMap[id]; !exist {
						ids = append(ids, id)
					}
				}

				if len(ids) > 0 {
					for i, id := range ids {
						if i > 0 {
							filter += ", "
						}
						filter += fmt.Sprintf(`'%s'`, id)
					}
					filter += ")"

					req.AdditionalQuery += filter + req.Order
					rows, err := req.Conn.Query(ctx, req.AdditionalQuery)
					if err != nil {
						return req.Result, nil
					}

					defer rows.Close()

					for rows.Next() {
						var (
							data interface{}
							temp = make(map[string]interface{})
						)

						values, err := rows.Values()
						if err != nil {
							return req.Result, nil
						}

						for i, value := range values {
							temp[rows.FieldDescriptions()[i].Name] = value
							data = temp["data"]
						}

						req.Result = append(req.Result, data)
					}

				}
			}
		}
	}

	return req.Result, nil
}
