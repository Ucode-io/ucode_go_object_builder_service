package postgres_test

import (
	"context"
	"fmt"
	"testing"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/pkg/helper"

	"github.com/stretchr/testify/assert"
)

func TestAgGridTree(t *testing.T) {
	treeRequest := map[string]interface{}{
		"startRow": 0,
		"endRow":   100,
		"rowGroupCols": []map[string]interface{}{
			{
				"id":          "country",
				"displayName": "Country",
				"field":       "country",
			},
		},
		"valueCols": []map[string]interface{}{
			{
				"id":          "gold",
				"aggFunc":     "sum",
				"displayName": "Gold",
				"field":       "gold",
			},
			{
				"id":          "silver",
				"aggFunc":     "sum",
				"displayName": "Silver",
				"field":       "silver",
			},
			{
				"id":          "bronze",
				"aggFunc":     "sum",
				"displayName": "Bronze",
				"field":       "bronze",
			},
		},
		"groupKeys":   []string{},
		"filterModel": map[string]interface{}{},
		"sortModel":   []map[string]interface{}{},
		"pivotMode":   false,
	}
	resp, err := helper.ConvertMapToStruct(treeRequest)
	assert.NoError(t, err)
	request := &nb.CommonMessage{
		TableSlug: "olympic",
		ProjectId: "f0259839-c2fc-44e8-af90-1a6aa7ba43f7",
		Data:      resp,
	}

	res, err := strg.ObjectBuilder().AgGridTree(context.Background(), request)
	assert.NoError(t, err)
	fmt.Println(res)
}
