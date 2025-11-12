package postgres_test

import (
	"context"
	"log"
	"testing"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/pkg/helper"

	"github.com/stretchr/testify/assert"
)

func GetListV2Test(t *testing.T, slug string) {
	req := &nb.CommonMessage{
		TableSlug: slug,
		ProjectId: projectId,
	}

	_, err := strg.ObjectBuilder().GetListV2(context.Background(), req)
	assert.NoError(t, err)
}

func GetListAggregationTest(t *testing.T, slug string) {

	data := map[string]any{
		"operation": "SELECT",
		"columns":   []string{"first_name", "last_name"},
		"table":     slug,
	}

	dataStruct, err := helper.ConvertMapToStruct(data)
	assert.NoError(t, err)

	req := &nb.CommonMessage{
		TableSlug: slug,
		Data:      dataStruct,
		ProjectId: projectId,
	}

	_, err = strg.ObjectBuilder().GetListAggregation(context.Background(), req)
	assert.NoError(t, err)
}

func GetTableDetailsTest(t *testing.T, slug string) {

	data := map[string]any{
		"role_id_from_token": "5381a752-0652-4da2-acfc-0dea5082a21e",
	}

	dataStruct, err := helper.ConvertMapToStruct(data)
	assert.NoError(t, err)

	req := &nb.CommonMessage{
		TableSlug: slug,
		Data:      dataStruct,
		ProjectId: projectId,
	}

	resp, err := strg.ObjectBuilder().GetTableDetails(context.Background(), req)
	assert.NoError(t, err)

	log.Println("RESPONSE:", resp)
}
