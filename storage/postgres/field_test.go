package postgres_test

import (
	"context"
	"testing"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"

	"github.com/stretchr/testify/assert"
)

func createField(t *testing.T) string {

	usage := &nb.CreateFieldRequest{
		Default:       "",
		Type:          "SINGLE_LINE",
		Index:         "string",
		Label:         "Name",
		Slug:          "name",
		TableId:       "66bfeb4e-8dc7-4214-bf0e-f4e72d07a191",
		IsVisible:     true,
		AutofillTable: "",
		AutofillField: "",
		Automatic:     false,
	}

	field, err := strg.Field().Create(context.Background(), usage)

	assert.NoError(t, err)
	assert.NotEmpty(t, field)

	return field.Id
}

func TestCreateField(t *testing.T) {
	createField(t)
}

// ? It's worked, but idk how to structpb type equals
func TestFieldGetSingle(t *testing.T) {
	expectedTable := &nb.Field{
		Id:            "120accdd-c7da-4bd8-bce6-2a63eda62883",
		Default:       "",
		Type:          "SINGLE_LINE",
		Index:         "string",
		Label:         "Name",
		Slug:          "name",
		TableId:       "66bfeb4e-8dc7-4214-bf0e-f4e72d07a191",
		IsVisible:     true,
		AutofillTable: "",
		AutofillField: "",
		Automatic:     false,
	}

	resp, err := strg.Field().GetByID(context.Background(), &nb.FieldPrimaryKey{Id: expectedTable.Id, ProjectId: ""})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, expectedTable, resp)
}
