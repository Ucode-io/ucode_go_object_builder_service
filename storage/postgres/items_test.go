package postgres_test

import (
	"context"
	"fmt"
	"testing"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/pkg/helper"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func createTable(t *testing.T, slug, label string) string {
	req := &nb.CreateTableRequest{
		Label:     label,
		Slug:      slug,
		ProjectId: projectId,
	}

	resp, err := strg.Table().Create(context.Background(), req)
	assert.NoError(t, err)
	return resp.Id
}

func createRelation(t *testing.T, tableFrom, tableTo string) {
	req := &nb.CreateRelationRequest{
		TableFrom:       tableFrom,
		TableTo:         tableTo,
		Type:            "Many2One",
		ViewFields:      []string{},
		RelationFieldId: uuid.New().String(),
		ProjectId:       projectId,
	}

	_, err = strg.Relation().Create(context.Background(), req)
	assert.NoError(t, err)
}

func createFields(t *testing.T, tableId string, fields []map[string]any) {
	for _, f := range fields {
		req := &nb.CreateFieldRequest{
			Id:            uuid.New().String(),
			ProjectId:     projectId,
			TableId:       tableId,
			Slug:          fmt.Sprintf("%v", f["slug"]),
			Label:         fmt.Sprintf("%v", f["label"]),
			Type:          fmt.Sprintf("%v", f["type"]),
			Index:         fmt.Sprintf("%v", f["index"]),
			IsVisible:     f["is_visible"].(bool),
			Unique:        f["unique"].(bool),
			AutofillField: fmt.Sprintf("%v", f["automatic"]),
		}

		_, err := strg.Field().Create(context.Background(), req)
		assert.NoError(t, err)
	}
}

func createItem(t *testing.T, tableSlug string, dataMap map[string]any) string {
	id := uuid.NewString()
	if dataMap == nil {
		dataMap = map[string]any{
			"from_auth_service": false,
			"guid":              id,
			"first_name":        fakeData.Person().FirstName(),
			"last_name":         fakeData.Person().LastName(),
		}
	} else {
		if _, ok := dataMap["guid"]; !ok {
			dataMap["guid"] = id
		}
	}

	data, err := helper.ConvertMapToStruct(dataMap)
	assert.NoError(t, err)

	req := &nb.CommonMessage{
		ProjectId: projectId,
		TableSlug: tableSlug,
		Data:      data,
	}

	_, err = strg.Items().Create(context.Background(), req)
	assert.NoError(t, err)

	return id
}

func getSingleItem(t *testing.T, tableSlug, id string) {

	dataMap := map[string]any{
		"from_auth_service": false,
		"id":                id,
	}

	data, err := helper.ConvertMapToStruct(dataMap)
	assert.NoError(t, err)

	req := &nb.CommonMessage{
		ProjectId: projectId,
		TableSlug: tableSlug,
		Data:      data,
	}

	_, err = strg.Items().GetSingle(context.Background(), req)
	assert.NoError(t, err)
}

func updateItem(t *testing.T, tableSlug string, payload map[string]any) {
	data, err := helper.ConvertMapToStruct(payload)
	assert.NoError(t, err)

	_, err = strg.Items().Update(
		context.Background(),
		&nb.CommonMessage{
			ProjectId: projectId,
			TableSlug: tableSlug,
			Data:      data,
		},
	)
	assert.NoError(t, err)
}

func deleteMany(t *testing.T, tableSlug string, ids []string) {
	dataMap := map[string]any{
		"from_auth_service": false,
		"ids":               ids,
	}

	data, err := helper.ConvertMapToStruct(dataMap)
	assert.NoError(t, err)

	_, err = strg.Items().DeleteMany(context.Background(),
		&nb.CommonMessage{
			ProjectId: projectId,
			TableSlug: tableSlug,
			Data:      data,
		},
	)
	assert.NoError(t, err)
}

func deleteSingle(t *testing.T, tableSlug, id string) {
	dataMap := map[string]any{
		"from_auth_service": false,
		"id":                id,
	}

	data, err := helper.ConvertMapToStruct(dataMap)
	assert.NoError(t, err)

	_, err = strg.Items().Delete(context.Background(),
		&nb.CommonMessage{
			ProjectId: projectId,
			TableSlug: tableSlug,
			Data:      data,
		})
	assert.NoError(t, err)
}

func upsertMany(t *testing.T, tableSlug string, fieldSlug string, fields []string, objects []map[string]any) {
	dataMap := map[string]any{
		"field_slug": fieldSlug,
		"fields":     fields,
		"objects":    objects,
	}

	data, err := helper.ConvertMapToStruct(dataMap)
	assert.NoError(t, err)

	err = strg.Items().UpsertMany(context.Background(),
		&nb.CommonMessage{
			ProjectId: projectId,
			TableSlug: tableSlug,
			Data:      data,
		},
	)
	assert.NoError(t, err)
}

func deleteTable(t *testing.T, id string) {
	err = strg.Table().Delete(context.Background(), &nb.TablePrimaryKey{Id: id, ProjectId: projectId})
	assert.NoError(t, err)
}

var (
	table1Slug = "test_table_1"
	table2Slug = "test_table_2"

	table1Label = "Test Table 1"
	table2Label = "Related Table"

	table1Fields = []map[string]any{
		{
			"slug":       "first_name",
			"label":      "First Name",
			"type":       "string",
			"index":      "btree",
			"is_visible": true,
			"unique":     false,
			"automatic":  false,
			"required":   true,
		},
		{
			"slug":       "last_name",
			"label":      "Last Name",
			"type":       "string",
			"index":      "btree",
			"is_visible": true,
			"unique":     false,
			"automatic":  false,
			"required":   true,
		},
	}

	table2Fields = []map[string]any{
		{
			"slug":       "name",
			"label":      "Name",
			"type":       "string",
			"index":      "btree",
			"is_visible": true,
			"unique":     false,
			"automatic":  false,
			"required":   true,
		},
	}
)

func TestItemsFlow(t *testing.T) {
	
	// 1) Create main table
	mainTableId := createTable(t, table1Slug, table1Label)

	// 2) Create fields for main table
	createFields(t, mainTableId, table1Fields)

	// 3) Create related table
	relatedTableId := createTable(t, table2Slug, table2Label)

	// 4) Create fields for related table
	createFields(t, relatedTableId, table2Fields)

	// 5) Create relation for main table
	createRelation(t, table1Slug, table2Slug)

	// 5) Create related item (will be referenced via author_id)
	relatedId := createItem(t, table2Slug, map[string]any{
		"from_auth_service": false,
		"name":              fakeData.Person().Name(),
	})

	// 6) Create main item with relation to relatedId
	mainId := createItem(t, table1Slug, map[string]any{
		"from_auth_service": false,
		"first_name":        fakeData.Person().FirstName(),
		"last_name":         fakeData.Person().LastName(),
		"test_table_2_id":   relatedId,
	})

	getSingleItem(t, table1Slug, mainId)

	// 8) Update main item
	updateItem(t, table1Slug, map[string]any{
		"from_auth_service": false,
		"guid":              mainId,
		"first_name":        fakeData.Person().FirstName(),
	})

	// 9) MultipleUpdate: update existing and add new
	existingId := mainId
	newId := uuid.NewString()
	{
		objects := []map[string]any{
			{
				"from_auth_service": false,
				"guid":              existingId,
				"first_name":        fakeData.Person().FirstName(),
				"last_name":         fakeData.Person().LastName(),
				"is_new":            false,
			},
			{
				"from_auth_service": false,
				"guid":              newId,
				"first_name":        fakeData.Person().FirstName(),
				"last_name":         fakeData.Person().LastName(),
				"is_new":            true,
			},
		}
		dataMap := map[string]any{"objects": objects}
		data, err := helper.ConvertMapToStruct(dataMap)
		assert.NoError(t, err)

		_, err = strg.Items().MultipleUpdate(context.Background(), &nb.CommonMessage{ProjectId: projectId, TableSlug: table1Slug, Data: data})
		assert.NoError(t, err)
	}

	// 10) UpsertMany
	existingIdForUpsert := existingId
	upsertNewId := uuid.NewString()

	{
		upsertObjects := []map[string]any{
			{
				"guid":       existingIdForUpsert,
				"first_name": fakeData.Person().FirstName(),
				"last_name":  fakeData.Person().LastName(),
			},
			{
				"guid":       upsertNewId,
				"first_name": fakeData.Person().FirstName(),
				"last_name":  fakeData.Person().LastName(),
			},
		}
		upsertMany(t, table1Slug, "guid", []string{"guid", "first_name", "last_name"}, upsertObjects)
	}

	// 11) DeleteMany (remove existing and the one created by MultipleUpdate)
	deleteMany(t, table1Slug, []string{existingId, newId})

	// 12) Delete single (remove the one created by UpsertMany)
	deleteSingle(t, table1Slug, upsertNewId)

	// 12) Delete single (delete from 2 table)
	deleteSingle(t, table2Slug, relatedId)

	deleteTable(t, mainTableId)
	deleteTable(t, relatedTableId)

}
