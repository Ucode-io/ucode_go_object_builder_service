package postgres_test

import (
	"context"
	"testing"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/pkg/helper"
)

func GetListV2Test(t *testing.T, slug string) error {
	req := &nb.CommonMessage{
		TableSlug: slug,
		ProjectId: projectId,
	}

	_, err := strg.ObjectBuilder().GetListV2(context.Background(), req)
	return err
}

func GetListAggregationTest(t *testing.T, slug string) error {

	data := map[string]any{
		"operation": "SELECT",
		"columns":   []string{"guid", "first_name"},
		"table":     slug,
	}

	dataStruct, err := helper.ConvertMapToStruct(data)
	if err != nil {
		return err
	}

	req := &nb.CommonMessage{
		TableSlug: slug,
		Data:      dataStruct,
		ProjectId: projectId,
	}

	_, err = strg.ObjectBuilder().GetListAggregation(context.Background(), req)
	return err
}

func GetTableDetailsTest(t *testing.T, slug string) error {

	data := map[string]any{
		"role_id_from_token": roleId,
	}

	dataStruct, err := helper.ConvertMapToStruct(data)
	if err != nil {
		return err
	}

	req := &nb.CommonMessage{
		TableSlug: slug,
		Data:      dataStruct,
		ProjectId: projectId,
	}

	_, err = strg.ObjectBuilder().GetTableDetails(context.Background(), req)
	return err
}
